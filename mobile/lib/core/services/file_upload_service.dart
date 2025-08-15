import 'dart:io';
import 'dart:convert';
import 'dart:typed_data';
import 'package:dio/dio.dart';
import 'package:path/path.dart' as path;
import 'package:mime/mime.dart';
import 'package:crypto/crypto.dart';
import 'package:logger/logger.dart';
import '../network/api_client.dart';
import '../storage/database_service.dart';
import '../utils/app_exceptions.dart';
import '../config/app_config.dart';

class FileUploadService {
  static FileUploadService? _instance;
  final Logger _logger = Logger();
  final Dio _dio = ApiClient().dio;
  final DatabaseService _databaseService = DatabaseService();

  FileUploadService._internal();

  factory FileUploadService() {
    _instance ??= FileUploadService._internal();
    return _instance!;
  }

  /// Uploads a file to the backend storage service
  Future<UploadResult> uploadFile(
    String filePath, {
    String? customFileName,
    String? folder,
    Map<String, String>? metadata,
    Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  }) async {
    try {
      final File file = File(filePath);
      if (!await file.exists()) {
        throw FileUploadException('File does not exist: $filePath');
      }

      final String fileName = customFileName ?? path.basename(filePath);
      final String mimeType = lookupMimeType(filePath) ?? 'application/octet-stream';
      final int fileSize = await file.length();
      
      _logger.i('Starting upload: $fileName (${_formatBytes(fileSize)})');

      // Calculate file hash for integrity check
      final String fileHash = await _calculateFileHash(file);

      // Create upload request
      final FormData formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          filePath,
          filename: fileName,
          contentType: DioMediaType.parse(mimeType),
        ),
        'folder': folder ?? 'uploads',
        'metadata': jsonEncode(metadata ?? {}),
        'hash': fileHash,
      });

      // Upload file
      final Response response = await _dio.post(
        '/storage/upload',
        data: formData,
        onSendProgress: onProgress,
        cancelToken: cancelToken,
        options: Options(
          headers: {
            'Content-Type': 'multipart/form-data',
          },
          sendTimeout: const Duration(minutes: 10),
          receiveTimeout: const Duration(minutes: 5),
        ),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        final uploadData = response.data as Map<String, dynamic>;
        
        final UploadResult result = UploadResult(
          id: uploadData['id'] as String,
          url: uploadData['url'] as String,
          fileName: uploadData['fileName'] as String,
          fileSize: uploadData['fileSize'] as int,
          mimeType: uploadData['mimeType'] as String,
          folder: uploadData['folder'] as String?,
          hash: uploadData['hash'] as String?,
          uploadedAt: DateTime.parse(uploadData['uploadedAt'] as String),
          metadata: uploadData['metadata'] as Map<String, dynamic>?,
        );

        _logger.i('Upload successful: ${result.fileName} -> ${result.url}');
        return result;
      } else {
        throw FileUploadException('Upload failed with status: ${response.statusCode}');
      }
    } on DioException catch (e) {
      _logger.e('Upload failed: ${e.message}');
      if (e.type == DioExceptionType.cancel) {
        throw FileUploadException('Upload cancelled');
      } else if (e.type == DioExceptionType.sendTimeout) {
        throw FileUploadException('Upload timeout');
      } else {
        throw FileUploadException('Upload failed: ${e.message}');
      }
    } catch (e) {
      _logger.e('Upload error: $e');
      throw FileUploadException('Upload failed: $e');
    }
  }

  /// Uploads multiple files
  Future<List<UploadResult>> uploadFiles(
    List<String> filePaths, {
    String? folder,
    Map<String, String>? metadata,
    Function(int current, int total, String fileName)? onFileProgress,
    Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  }) async {
    final List<UploadResult> results = [];
    
    for (int i = 0; i < filePaths.length; i++) {
      final String filePath = filePaths[i];
      final String fileName = path.basename(filePath);
      
      onFileProgress?.call(i + 1, filePaths.length, fileName);
      
      try {
        final result = await uploadFile(
          filePath,
          folder: folder,
          metadata: metadata,
          onProgress: onProgress,
          cancelToken: cancelToken,
        );
        results.add(result);
      } catch (e) {
        _logger.e('Failed to upload $fileName: $e');
        // Continue with other files
      }
    }
    
    return results;
  }

  /// Queues a file for background upload when connection is available
  Future<void> queueFileUpload(
    String filePath, {
    String? customFileName,
    String? folder,
    String? relatedTable,
    String? relatedId,
    Map<String, String>? metadata,
    int priority = 1,
  }) async {
    try {
      final File file = File(filePath);
      if (!await file.exists()) {
        throw FileUploadException('File does not exist: $filePath');
      }

      await _databaseService.insert('file_upload_queue', {
        'file_path': filePath,
        'file_type': lookupMimeType(filePath) ?? 'application/octet-stream',
        'related_table': relatedTable,
        'related_id': relatedId,
        'metadata': jsonEncode({
          'custom_file_name': customFileName,
          'folder': folder,
          'upload_metadata': metadata,
        }),
        'created_at': DateTime.now().toIso8601String(),
        'priority': priority,
      });

      _logger.i('File queued for upload: ${path.basename(filePath)}');
    } catch (e) {
      _logger.e('Failed to queue file upload: $e');
      throw FileUploadException('Failed to queue file upload: $e');
    }
  }

  /// Processes the upload queue
  Future<void> processUploadQueue() async {
    try {
      final List<Map<String, dynamic>> queuedUploads = await _databaseService.query(
        'file_upload_queue',
        where: 'is_uploaded = ? AND attempts < ?',
        whereArgs: [0, 3],
        orderBy: 'priority DESC, created_at ASC',
        limit: 10,
      );

      for (final queueItem in queuedUploads) {
        try {
          await _processQueuedUpload(queueItem);
        } catch (e) {
          _logger.e('Failed to process queued upload: $e');
          await _updateQueueItemError(queueItem['id'] as int, e.toString());
        }
      }
    } catch (e) {
      _logger.e('Failed to process upload queue: $e');
    }
  }

  /// Downloads a file from URL
  Future<String> downloadFile(
    String url, {
    String? savePath,
    Function(int received, int total)? onProgress,
    CancelToken? cancelToken,
  }) async {
    try {
      final String fileName = Uri.parse(url).pathSegments.last;
      final String finalPath = savePath ?? await _getDownloadPath(fileName);
      
      _logger.i('Downloading file: $fileName');

      await _dio.download(
        url,
        finalPath,
        onReceiveProgress: onProgress,
        cancelToken: cancelToken,
        options: Options(
          receiveTimeout: const Duration(minutes: 10),
        ),
      );

      _logger.i('Download complete: $finalPath');
      return finalPath;
    } on DioException catch (e) {
      _logger.e('Download failed: ${e.message}');
      if (e.type == DioExceptionType.cancel) {
        throw FileUploadException('Download cancelled');
      } else {
        throw FileUploadException('Download failed: ${e.message}');
      }
    } catch (e) {
      _logger.e('Download error: $e');
      throw FileUploadException('Download failed: $e');
    }
  }

  /// Deletes a file from storage
  Future<bool> deleteFile(String fileId) async {
    try {
      final Response response = await _dio.delete('/storage/files/$fileId');
      return response.statusCode == 200 || response.statusCode == 204;
    } on DioException catch (e) {
      _logger.e('Failed to delete file: ${e.message}');
      return false;
    }
  }

  /// Gets file information
  Future<FileInfo?> getFileInfo(String fileId) async {
    try {
      final Response response = await _dio.get('/storage/files/$fileId');
      
      if (response.statusCode == 200) {
        final data = response.data as Map<String, dynamic>;
        return FileInfo.fromJson(data);
      }
      return null;
    } on DioException catch (e) {
      _logger.e('Failed to get file info: ${e.message}');
      return null;
    }
  }

  /// Processes a single queued upload
  Future<void> _processQueuedUpload(Map<String, dynamic> queueItem) async {
    final int id = queueItem['id'] as int;
    final String filePath = queueItem['file_path'] as String;
    final Map<String, dynamic> metadata = jsonDecode(queueItem['metadata'] as String);

    // Update attempts
    await _databaseService.update(
      'file_upload_queue',
      {
        'attempts': (queueItem['attempts'] as int) + 1,
        'last_attempt_at': DateTime.now().toIso8601String(),
      },
      where: 'id = ?',
      whereArgs: [id],
    );

    // Attempt upload
    final UploadResult result = await uploadFile(
      filePath,
      customFileName: metadata['custom_file_name'] as String?,
      folder: metadata['folder'] as String?,
      metadata: metadata['upload_metadata'] as Map<String, String>?,
    );

    // Mark as uploaded
    await _databaseService.update(
      'file_upload_queue',
      {
        'is_uploaded': 1,
        'uploaded_at': DateTime.now().toIso8601String(),
        'upload_url': result.url,
      },
      where: 'id = ?',
      whereArgs: [id],
    );

    // Update related record if specified
    final String? relatedTable = queueItem['related_table'] as String?;
    final String? relatedId = queueItem['related_id'] as String?;
    
    if (relatedTable != null && relatedId != null) {
      await _updateRelatedRecord(relatedTable, relatedId, result);
    }
  }

  /// Updates queue item with error
  Future<void> _updateQueueItemError(int id, String error) async {
    await _databaseService.update(
      'file_upload_queue',
      {
        'error_message': error,
        'last_attempt_at': DateTime.now().toIso8601String(),
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  /// Updates related record with upload result
  Future<void> _updateRelatedRecord(
    String tableName,
    String recordId,
    UploadResult result,
  ) async {
    try {
      if (tableName == 'job_photos') {
        await _databaseService.update(
          tableName,
          {
            'url': result.url,
            'is_uploaded': 1,
            'uploaded_at': result.uploadedAt.toIso8601String(),
          },
          where: 'id = ?',
          whereArgs: [recordId],
        );
      }
    } catch (e) {
      _logger.e('Failed to update related record: $e');
    }
  }

  /// Calculates file hash for integrity verification
  Future<String> _calculateFileHash(File file) async {
    final Uint8List bytes = await file.readAsBytes();
    final Digest digest = sha256.convert(bytes);
    return digest.toString();
  }

  /// Gets download path for file
  Future<String> _getDownloadPath(String fileName) async {
    // Implementation depends on platform and requirements
    // For now, use a simple approach
    final Directory tempDir = Directory.systemTemp;
    return path.join(tempDir.path, fileName);
  }

  /// Formats bytes to human readable string
  String _formatBytes(int bytes) {
    const int kb = 1024;
    const int mb = kb * 1024;
    
    if (bytes >= mb) {
      return '${(bytes / mb).toStringAsFixed(1)} MB';
    } else if (bytes >= kb) {
      return '${(bytes / kb).toStringAsFixed(1)} KB';
    } else {
      return '$bytes bytes';
    }
  }

  /// Cleans up completed queue items
  Future<void> cleanupQueue() async {
    try {
      // Remove uploaded items older than 7 days
      final cutoffDate = DateTime.now().subtract(const Duration(days: 7));
      
      await _databaseService.delete(
        'file_upload_queue',
        where: 'is_uploaded = 1 AND uploaded_at < ?',
        whereArgs: [cutoffDate.toIso8601String()],
      );

      // Remove failed items after 3 attempts and older than 1 day
      final failedCutoffDate = DateTime.now().subtract(const Duration(days: 1));
      
      await _databaseService.delete(
        'file_upload_queue',
        where: 'attempts >= 3 AND last_attempt_at < ?',
        whereArgs: [failedCutoffDate.toIso8601String()],
      );

      _logger.i('Upload queue cleaned up');
    } catch (e) {
      _logger.e('Failed to cleanup upload queue: $e');
    }
  }

  /// Gets upload queue status
  Future<UploadQueueStatus> getQueueStatus() async {
    try {
      final List<Map<String, dynamic>> pending = await _databaseService.query(
        'file_upload_queue',
        where: 'is_uploaded = 0',
      );

      final List<Map<String, dynamic>> failed = await _databaseService.query(
        'file_upload_queue',
        where: 'attempts >= 3 AND is_uploaded = 0',
      );

      final List<Map<String, dynamic>> completed = await _databaseService.query(
        'file_upload_queue',
        where: 'is_uploaded = 1',
      );

      return UploadQueueStatus(
        pending: pending.length,
        failed: failed.length,
        completed: completed.length,
      );
    } catch (e) {
      _logger.e('Failed to get queue status: $e');
      return const UploadQueueStatus(pending: 0, failed: 0, completed: 0);
    }
  }
}

class UploadResult {
  final String id;
  final String url;
  final String fileName;
  final int fileSize;
  final String mimeType;
  final String? folder;
  final String? hash;
  final DateTime uploadedAt;
  final Map<String, dynamic>? metadata;

  const UploadResult({
    required this.id,
    required this.url,
    required this.fileName,
    required this.fileSize,
    required this.mimeType,
    this.folder,
    this.hash,
    required this.uploadedAt,
    this.metadata,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'url': url,
    'fileName': fileName,
    'fileSize': fileSize,
    'mimeType': mimeType,
    'folder': folder,
    'hash': hash,
    'uploadedAt': uploadedAt.toIso8601String(),
    'metadata': metadata,
  };

  factory UploadResult.fromJson(Map<String, dynamic> json) => UploadResult(
    id: json['id'] as String,
    url: json['url'] as String,
    fileName: json['fileName'] as String,
    fileSize: json['fileSize'] as int,
    mimeType: json['mimeType'] as String,
    folder: json['folder'] as String?,
    hash: json['hash'] as String?,
    uploadedAt: DateTime.parse(json['uploadedAt'] as String),
    metadata: json['metadata'] as Map<String, dynamic>?,
  );
}

class FileInfo {
  final String id;
  final String fileName;
  final int fileSize;
  final String mimeType;
  final String url;
  final String? folder;
  final DateTime createdAt;
  final Map<String, dynamic>? metadata;

  const FileInfo({
    required this.id,
    required this.fileName,
    required this.fileSize,
    required this.mimeType,
    required this.url,
    this.folder,
    required this.createdAt,
    this.metadata,
  });

  factory FileInfo.fromJson(Map<String, dynamic> json) => FileInfo(
    id: json['id'] as String,
    fileName: json['fileName'] as String,
    fileSize: json['fileSize'] as int,
    mimeType: json['mimeType'] as String,
    url: json['url'] as String,
    folder: json['folder'] as String?,
    createdAt: DateTime.parse(json['createdAt'] as String),
    metadata: json['metadata'] as Map<String, dynamic>?,
  );
}

class UploadQueueStatus {
  final int pending;
  final int failed;
  final int completed;

  const UploadQueueStatus({
    required this.pending,
    required this.failed,
    required this.completed,
  });

  int get total => pending + failed + completed;
}

class FileUploadException extends AppException {
  FileUploadException(String message) : super(message);
}