import 'dart:async';
import 'dart:io';
import 'dart:typed_data';
import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;
import 'package:logger/logger.dart';
import 'package:geolocator/geolocator.dart';
import '../utils/app_exceptions.dart';

class CameraService {
  static CameraService? _instance;
  final Logger _logger = Logger();
  final ImagePicker _imagePicker = ImagePicker();
  
  List<CameraDescription> _cameras = [];
  CameraController? _controller;
  bool _isInitialized = false;

  CameraService._internal();

  factory CameraService() {
    _instance ??= CameraService._internal();
    return _instance!;
  }

  List<CameraDescription> get cameras => _cameras;
  CameraController? get controller => _controller;
  bool get isInitialized => _isInitialized;

  Future<void> initialize() async {
    try {
      _cameras = await availableCameras();
      _logger.i('Found ${_cameras.length} cameras');
      _isInitialized = true;
    } catch (e) {
      _logger.e('Failed to initialize cameras: $e');
      throw CameraException('Failed to initialize cameras: $e');
    }
  }

  Future<bool> checkPermissions() async {
    final cameraStatus = await Permission.camera.status;
    final storageStatus = await Permission.storage.status;
    
    if (cameraStatus.isDenied || storageStatus.isDenied) {
      final results = await [
        Permission.camera,
        Permission.storage,
      ].request();
      
      return results[Permission.camera]!.isGranted && 
             results[Permission.storage]!.isGranted;
    }
    
    return cameraStatus.isGranted && storageStatus.isGranted;
  }

  Future<CameraController> initializeCamera({
    CameraDescription? camera,
    ResolutionPreset resolution = ResolutionPreset.high,
    bool enableAudio = false,
  }) async {
    if (!_isInitialized) {
      await initialize();
    }

    if (_cameras.isEmpty) {
      throw CameraException('No cameras available');
    }

    final selectedCamera = camera ?? _cameras.first;
    
    _controller = CameraController(
      selectedCamera,
      resolution,
      enableAudio: enableAudio,
      imageFormatGroup: ImageFormatGroup.jpeg,
    );

    await _controller!.initialize();
    _logger.i('Camera initialized: ${selectedCamera.name}');
    
    return _controller!;
  }

  Future<CapturedPhoto> takePicture({
    String? jobId,
    String? description,
    Map<String, dynamic>? metadata,
  }) async {
    if (_controller == null || !_controller!.value.isInitialized) {
      throw CameraException('Camera not initialized');
    }

    try {
      final XFile picture = await _controller!.takePicture();
      final File imageFile = File(picture.path);
      
      // Get location if available
      Position? position;
      try {
        position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
          timeLimit: const Duration(seconds: 5),
        );
      } catch (e) {
        _logger.w('Could not get location: $e');
      }

      // Create permanent file path
      final String fileName = _generateFileName(jobId);
      final String permanentPath = await _savePermanently(imageFile, fileName);

      return CapturedPhoto(
        localPath: permanentPath,
        originalPath: picture.path,
        fileName: fileName,
        fileSize: await imageFile.length(),
        width: _controller!.value.previewSize?.width.toInt(),
        height: _controller!.value.previewSize?.height.toInt(),
        latitude: position?.latitude,
        longitude: position?.longitude,
        capturedAt: DateTime.now(),
        jobId: jobId,
        description: description,
        metadata: metadata,
      );
    } catch (e) {
      _logger.e('Failed to take picture: $e');
      throw CameraException('Failed to take picture: $e');
    }
  }

  Future<CapturedPhoto> pickFromGallery({
    String? jobId,
    String? description,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      final XFile? pickedFile = await _imagePicker.pickImage(
        source: ImageSource.gallery,
        imageQuality: 85,
        maxWidth: 2048,
        maxHeight: 2048,
      );

      if (pickedFile == null) {
        throw CameraException('No image selected');
      }

      final File imageFile = File(pickedFile.path);
      
      // Create permanent file path
      final String fileName = _generateFileName(jobId);
      final String permanentPath = await _savePermanently(imageFile, fileName);

      return CapturedPhoto(
        localPath: permanentPath,
        originalPath: pickedFile.path,
        fileName: fileName,
        fileSize: await imageFile.length(),
        capturedAt: DateTime.now(),
        jobId: jobId,
        description: description,
        metadata: metadata,
      );
    } catch (e) {
      _logger.e('Failed to pick image from gallery: $e');
      throw CameraException('Failed to pick image from gallery: $e');
    }
  }

  Future<List<CapturedPhoto>> pickMultipleFromGallery({
    String? jobId,
    int maxImages = 10,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      final List<XFile> pickedFiles = await _imagePicker.pickMultiImage(
        imageQuality: 85,
        maxWidth: 2048,
        maxHeight: 2048,
      );

      if (pickedFiles.isEmpty) {
        throw CameraException('No images selected');
      }

      if (pickedFiles.length > maxImages) {
        throw CameraException('Too many images selected. Maximum is $maxImages');
      }

      final List<CapturedPhoto> photos = [];

      for (final XFile pickedFile in pickedFiles) {
        final File imageFile = File(pickedFile.path);
        
        // Create permanent file path
        final String fileName = _generateFileName(jobId);
        final String permanentPath = await _savePermanently(imageFile, fileName);

        photos.add(CapturedPhoto(
          localPath: permanentPath,
          originalPath: pickedFile.path,
          fileName: fileName,
          fileSize: await imageFile.length(),
          capturedAt: DateTime.now(),
          jobId: jobId,
          metadata: metadata,
        ));
      }

      return photos;
    } catch (e) {
      _logger.e('Failed to pick multiple images: $e');
      throw CameraException('Failed to pick multiple images: $e');
    }
  }

  Future<String> _savePermanently(File sourceFile, String fileName) async {
    try {
      final Directory appDir = await getApplicationDocumentsDirectory();
      final Directory photosDir = Directory(path.join(appDir.path, 'photos'));
      
      if (!await photosDir.exists()) {
        await photosDir.create(recursive: true);
      }

      final String targetPath = path.join(photosDir.path, fileName);
      await sourceFile.copy(targetPath);
      
      // Clean up temporary file if it's different from target
      if (sourceFile.path != targetPath) {
        await sourceFile.delete();
      }

      return targetPath;
    } catch (e) {
      _logger.e('Failed to save file permanently: $e');
      throw CameraException('Failed to save file: $e');
    }
  }

  String _generateFileName(String? jobId) {
    final timestamp = DateTime.now().millisecondsSinceEpoch;
    final jobPrefix = jobId != null ? '${jobId}_' : '';
    return '${jobPrefix}photo_$timestamp.jpg';
  }

  Future<void> switchCamera() async {
    if (_cameras.length < 2) {
      throw CameraException('Only one camera available');
    }

    final currentCameraIndex = _cameras.indexOf(_controller!.description);
    final nextCameraIndex = (currentCameraIndex + 1) % _cameras.length;
    
    await dispose();
    await initializeCamera(camera: _cameras[nextCameraIndex]);
  }

  Future<void> setFlashMode(FlashMode flashMode) async {
    if (_controller == null) {
      throw CameraException('Camera not initialized');
    }

    try {
      await _controller!.setFlashMode(flashMode);
    } catch (e) {
      _logger.e('Failed to set flash mode: $e');
      throw CameraException('Failed to set flash mode: $e');
    }
  }

  Future<void> setZoomLevel(double zoom) async {
    if (_controller == null) {
      throw CameraException('Camera not initialized');
    }

    try {
      final double maxZoom = await _controller!.getMaxZoomLevel();
      final double minZoom = await _controller!.getMinZoomLevel();
      final double clampedZoom = zoom.clamp(minZoom, maxZoom);
      
      await _controller!.setZoomLevel(clampedZoom);
    } catch (e) {
      _logger.e('Failed to set zoom level: $e');
      throw CameraException('Failed to set zoom level: $e');
    }
  }

  Future<void> setFocusPoint(Offset point) async {
    if (_controller == null) {
      throw CameraException('Camera not initialized');
    }

    try {
      await _controller!.setFocusPoint(point);
    } catch (e) {
      _logger.e('Failed to set focus point: $e');
      throw CameraException('Failed to set focus point: $e');
    }
  }

  Future<void> setExposurePoint(Offset point) async {
    if (_controller == null) {
      throw CameraException('Camera not initialized');
    }

    try {
      await _controller!.setExposurePoint(point);
    } catch (e) {
      _logger.e('Failed to set exposure point: $e');
      throw CameraException('Failed to set exposure point: $e');
    }
  }

  Future<void> dispose() async {
    if (_controller != null) {
      await _controller!.dispose();
      _controller = null;
    }
  }

  // Helper methods for file management
  Future<List<File>> getStoredPhotos({String? jobId}) async {
    try {
      final Directory appDir = await getApplicationDocumentsDirectory();
      final Directory photosDir = Directory(path.join(appDir.path, 'photos'));
      
      if (!await photosDir.exists()) {
        return [];
      }

      final List<FileSystemEntity> files = await photosDir.list().toList();
      final List<File> photoFiles = files
          .whereType<File>()
          .where((file) => file.path.toLowerCase().endsWith('.jpg') || 
                         file.path.toLowerCase().endsWith('.jpeg') ||
                         file.path.toLowerCase().endsWith('.png'))
          .toList();

      if (jobId != null) {
        return photoFiles
            .where((file) => path.basename(file.path).startsWith('${jobId}_'))
            .toList();
      }

      return photoFiles;
    } catch (e) {
      _logger.e('Failed to get stored photos: $e');
      return [];
    }
  }

  Future<void> deletePhoto(String filePath) async {
    try {
      final File file = File(filePath);
      if (await file.exists()) {
        await file.delete();
        _logger.i('Photo deleted: $filePath');
      }
    } catch (e) {
      _logger.e('Failed to delete photo: $e');
      throw CameraException('Failed to delete photo: $e');
    }
  }

  Future<int> getStoredPhotosSize() async {
    try {
      final photos = await getStoredPhotos();
      int totalSize = 0;
      
      for (final photo in photos) {
        totalSize += await photo.length();
      }
      
      return totalSize;
    } catch (e) {
      _logger.e('Failed to calculate photos size: $e');
      return 0;
    }
  }

  Future<void> cleanupOldPhotos({int maxAgeInDays = 30}) async {
    try {
      final photos = await getStoredPhotos();
      final cutoffDate = DateTime.now().subtract(Duration(days: maxAgeInDays));
      
      for (final photo in photos) {
        final fileStat = await photo.stat();
        if (fileStat.modified.isBefore(cutoffDate)) {
          await photo.delete();
          _logger.i('Deleted old photo: ${photo.path}');
        }
      }
    } catch (e) {
      _logger.e('Failed to cleanup old photos: $e');
    }
  }
}

class CapturedPhoto {
  final String localPath;
  final String originalPath;
  final String fileName;
  final int fileSize;
  final int? width;
  final int? height;
  final double? latitude;
  final double? longitude;
  final DateTime capturedAt;
  final String? jobId;
  final String? description;
  final Map<String, dynamic>? metadata;

  const CapturedPhoto({
    required this.localPath,
    required this.originalPath,
    required this.fileName,
    required this.fileSize,
    this.width,
    this.height,
    this.latitude,
    this.longitude,
    required this.capturedAt,
    this.jobId,
    this.description,
    this.metadata,
  });

  Map<String, dynamic> toJson() => {
    'localPath': localPath,
    'originalPath': originalPath,
    'fileName': fileName,
    'fileSize': fileSize,
    'width': width,
    'height': height,
    'latitude': latitude,
    'longitude': longitude,
    'capturedAt': capturedAt.toIso8601String(),
    'jobId': jobId,
    'description': description,
    'metadata': metadata,
  };

  factory CapturedPhoto.fromJson(Map<String, dynamic> json) => CapturedPhoto(
    localPath: json['localPath'] as String,
    originalPath: json['originalPath'] as String,
    fileName: json['fileName'] as String,
    fileSize: json['fileSize'] as int,
    width: json['width'] as int?,
    height: json['height'] as int?,
    latitude: json['latitude'] as double?,
    longitude: json['longitude'] as double?,
    capturedAt: DateTime.parse(json['capturedAt'] as String),
    jobId: json['jobId'] as String?,
    description: json['description'] as String?,
    metadata: json['metadata'] as Map<String, dynamic>?,
  );

  bool get hasLocation => latitude != null && longitude != null;
  
  String get sizeFormatted {
    const int kb = 1024;
    const int mb = kb * 1024;
    
    if (fileSize >= mb) {
      return '${(fileSize / mb).toStringAsFixed(1)} MB';
    } else if (fileSize >= kb) {
      return '${(fileSize / kb).toStringAsFixed(1)} KB';
    } else {
      return '$fileSize bytes';
    }
  }
}

class CameraException extends AppException {
  CameraException(String message) : super(message);
}