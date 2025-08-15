import 'dart:io';
import 'dart:typed_data';
import 'package:flutter_image_compress/flutter_image_compress.dart';
import 'package:image/image.dart' as img;
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;
import 'package:logger/logger.dart';
import '../utils/app_exceptions.dart';

class ImageCompressionService {
  static ImageCompressionService? _instance;
  final Logger _logger = Logger();

  ImageCompressionService._internal();

  factory ImageCompressionService() {
    _instance ??= ImageCompressionService._internal();
    return _instance!;
  }

  /// Compresses an image with various quality and size options
  Future<CompressedImage> compressImage(
    String inputPath, {
    int quality = 85,
    int? maxWidth,
    int? maxHeight,
    int? maxSizeKB,
    String? outputPath,
    ImageFormat format = ImageFormat.jpeg,
    bool preserveExif = false,
  }) async {
    try {
      final File inputFile = File(inputPath);
      if (!await inputFile.exists()) {
        throw ImageCompressionException('Input file does not exist: $inputPath');
      }

      final int originalSize = await inputFile.length();
      _logger.i('Compressing image: ${path.basename(inputPath)} (${_formatBytes(originalSize)})');

      // Generate output path if not provided
      final String finalOutputPath = outputPath ?? await _generateOutputPath(inputPath, format);

      Uint8List? compressedBytes;

      // Try flutter_image_compress first for better performance
      try {
        compressedBytes = await FlutterImageCompress.compressWithFile(
          inputPath,
          minWidth: maxWidth ?? 1920,
          minHeight: maxHeight ?? 1080,
          quality: quality,
          format: _getCompressFormat(format),
          keepExif: preserveExif,
        );
      } catch (e) {
        _logger.w('Flutter compression failed, falling back to dart:ui: $e');
        compressedBytes = await _compressWithDartImage(
          inputPath,
          quality: quality,
          maxWidth: maxWidth,
          maxHeight: maxHeight,
          format: format,
        );
      }

      if (compressedBytes == null) {
        throw ImageCompressionException('Failed to compress image');
      }

      // If max size is specified and we're still over, reduce quality iteratively
      if (maxSizeKB != null) {
        compressedBytes = await _compressToTargetSize(
          inputPath,
          maxSizeKB * 1024,
          quality,
          maxWidth,
          maxHeight,
          format,
          preserveExif,
        );
      }

      // Save compressed image
      final File outputFile = File(finalOutputPath);
      await outputFile.writeAsBytes(compressedBytes);

      final int compressedSize = compressedBytes.length;
      final double compressionRatio = (1.0 - (compressedSize / originalSize)) * 100;

      _logger.i('Compression complete: ${_formatBytes(compressedSize)} (${compressionRatio.toStringAsFixed(1)}% reduction)');

      return CompressedImage(
        originalPath: inputPath,
        compressedPath: finalOutputPath,
        originalSize: originalSize,
        compressedSize: compressedSize,
        compressionRatio: compressionRatio,
        quality: quality,
        format: format,
      );
    } catch (e) {
      _logger.e('Image compression failed: $e');
      throw ImageCompressionException('Failed to compress image: $e');
    }
  }

  /// Compresses multiple images
  Future<List<CompressedImage>> compressImages(
    List<String> inputPaths, {
    int quality = 85,
    int? maxWidth,
    int? maxHeight,
    int? maxSizeKB,
    ImageFormat format = ImageFormat.jpeg,
    bool preserveExif = false,
    Function(int current, int total)? onProgress,
  }) async {
    final List<CompressedImage> results = [];
    
    for (int i = 0; i < inputPaths.length; i++) {
      onProgress?.call(i + 1, inputPaths.length);
      
      try {
        final compressed = await compressImage(
          inputPaths[i],
          quality: quality,
          maxWidth: maxWidth,
          maxHeight: maxHeight,
          maxSizeKB: maxSizeKB,
          format: format,
          preserveExif: preserveExif,
        );
        results.add(compressed);
      } catch (e) {
        _logger.e('Failed to compress image ${inputPaths[i]}: $e');
        // Continue with other images
      }
    }
    
    return results;
  }

  /// Resizes an image to specific dimensions
  Future<String> resizeImage(
    String inputPath, {
    required int width,
    required int height,
    bool maintainAspectRatio = true,
    String? outputPath,
    ImageFormat format = ImageFormat.jpeg,
    int quality = 90,
  }) async {
    try {
      final File inputFile = File(inputPath);
      if (!await inputFile.exists()) {
        throw ImageCompressionException('Input file does not exist: $inputPath');
      }

      // Generate output path if not provided
      final String finalOutputPath = outputPath ?? await _generateOutputPath(inputPath, format);

      Uint8List? resizedBytes = await FlutterImageCompress.compressWithFile(
        inputPath,
        minWidth: width,
        minHeight: height,
        quality: quality,
        format: _getCompressFormat(format),
      );

      if (resizedBytes == null) {
        throw ImageCompressionException('Failed to resize image');
      }

      // Save resized image
      final File outputFile = File(finalOutputPath);
      await outputFile.writeAsBytes(resizedBytes);

      return finalOutputPath;
    } catch (e) {
      _logger.e('Image resize failed: $e');
      throw ImageCompressionException('Failed to resize image: $e');
    }
  }

  /// Creates thumbnails for images
  Future<String> createThumbnail(
    String inputPath, {
    int size = 200,
    String? outputPath,
    int quality = 80,
  }) async {
    return await resizeImage(
      inputPath,
      width: size,
      height: size,
      maintainAspectRatio: true,
      outputPath: outputPath,
      quality: quality,
    );
  }

  /// Gets image information without loading the full image
  Future<ImageInfo> getImageInfo(String imagePath) async {
    try {
      final File imageFile = File(imagePath);
      if (!await imageFile.exists()) {
        throw ImageCompressionException('Image file does not exist: $imagePath');
      }

      final Uint8List bytes = await imageFile.readAsBytes();
      final img.Image? image = img.decodeImage(bytes);

      if (image == null) {
        throw ImageCompressionException('Failed to decode image');
      }

      return ImageInfo(
        width: image.width,
        height: image.height,
        fileSize: bytes.length,
        format: _detectImageFormat(imagePath),
        hasTransparency: image.hasAlpha,
      );
    } catch (e) {
      _logger.e('Failed to get image info: $e');
      throw ImageCompressionException('Failed to get image info: $e');
    }
  }

  /// Compresses image to target size iteratively
  Future<Uint8List> _compressToTargetSize(
    String inputPath,
    int targetSizeBytes,
    int initialQuality,
    int? maxWidth,
    int? maxHeight,
    ImageFormat format,
    bool preserveExif,
  ) async {
    int quality = initialQuality;
    Uint8List? compressedBytes;
    
    while (quality > 10) {
      compressedBytes = await FlutterImageCompress.compressWithFile(
        inputPath,
        minWidth: maxWidth ?? 1920,
        minHeight: maxHeight ?? 1080,
        quality: quality,
        format: _getCompressFormat(format),
        keepExif: preserveExif,
      );

      if (compressedBytes != null && compressedBytes.length <= targetSizeBytes) {
        break;
      }

      quality -= 10;
    }

    if (compressedBytes == null) {
      throw ImageCompressionException('Failed to compress to target size');
    }

    return compressedBytes;
  }

  /// Fallback compression using dart:ui image library
  Future<Uint8List> _compressWithDartImage(
    String inputPath, {
    required int quality,
    int? maxWidth,
    int? maxHeight,
    required ImageFormat format,
  }) async {
    final File inputFile = File(inputPath);
    final Uint8List bytes = await inputFile.readAsBytes();
    
    img.Image? image = img.decodeImage(bytes);
    if (image == null) {
      throw ImageCompressionException('Failed to decode image with dart:ui');
    }

    // Resize if needed
    if (maxWidth != null || maxHeight != null) {
      image = img.copyResize(
        image,
        width: maxWidth,
        height: maxHeight,
        maintainAspect: true,
      );
    }

    // Encode with compression
    Uint8List compressedBytes;
    switch (format) {
      case ImageFormat.jpeg:
        compressedBytes = Uint8List.fromList(img.encodeJpg(image, quality: quality));
        break;
      case ImageFormat.png:
        compressedBytes = Uint8List.fromList(img.encodePng(image));
        break;
      case ImageFormat.webp:
        compressedBytes = Uint8List.fromList(img.encodeWebP(image, quality: quality));
        break;
    }

    return compressedBytes;
  }

  /// Generates output path for compressed image
  Future<String> _generateOutputPath(String inputPath, ImageFormat format) async {
    final Directory tempDir = await getTemporaryDirectory();
    final String fileName = path.basenameWithoutExtension(inputPath);
    final String extension = _getFileExtension(format);
    final String timestamp = DateTime.now().millisecondsSinceEpoch.toString();
    
    return path.join(tempDir.path, '${fileName}_compressed_$timestamp.$extension');
  }

  /// Converts ImageFormat to CompressFormat
  CompressFormat _getCompressFormat(ImageFormat format) {
    switch (format) {
      case ImageFormat.jpeg:
        return CompressFormat.jpeg;
      case ImageFormat.png:
        return CompressFormat.png;
      case ImageFormat.webp:
        return CompressFormat.webp;
    }
  }

  /// Gets file extension for format
  String _getFileExtension(ImageFormat format) {
    switch (format) {
      case ImageFormat.jpeg:
        return 'jpg';
      case ImageFormat.png:
        return 'png';
      case ImageFormat.webp:
        return 'webp';
    }
  }

  /// Detects image format from file path
  ImageFormat _detectImageFormat(String filePath) {
    final String extension = path.extension(filePath).toLowerCase();
    switch (extension) {
      case '.jpg':
      case '.jpeg':
        return ImageFormat.jpeg;
      case '.png':
        return ImageFormat.png;
      case '.webp':
        return ImageFormat.webp;
      default:
        return ImageFormat.jpeg;
    }
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

  /// Cleans up temporary compressed files
  Future<void> cleanupTempFiles() async {
    try {
      final Directory tempDir = await getTemporaryDirectory();
      final List<FileSystemEntity> files = await tempDir.list().toList();
      
      for (final file in files) {
        if (file is File && file.path.contains('_compressed_')) {
          await file.delete();
        }
      }
      
      _logger.i('Cleaned up temporary compressed files');
    } catch (e) {
      _logger.e('Failed to cleanup temp files: $e');
    }
  }
}

class CompressedImage {
  final String originalPath;
  final String compressedPath;
  final int originalSize;
  final int compressedSize;
  final double compressionRatio;
  final int quality;
  final ImageFormat format;

  const CompressedImage({
    required this.originalPath,
    required this.compressedPath,
    required this.originalSize,
    required this.compressedSize,
    required this.compressionRatio,
    required this.quality,
    required this.format,
  });

  String get originalSizeFormatted => _formatBytes(originalSize);
  String get compressedSizeFormatted => _formatBytes(compressedSize);
  
  static String _formatBytes(int bytes) {
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

  Map<String, dynamic> toJson() => {
    'originalPath': originalPath,
    'compressedPath': compressedPath,
    'originalSize': originalSize,
    'compressedSize': compressedSize,
    'compressionRatio': compressionRatio,
    'quality': quality,
    'format': format.name,
  };

  factory CompressedImage.fromJson(Map<String, dynamic> json) => CompressedImage(
    originalPath: json['originalPath'] as String,
    compressedPath: json['compressedPath'] as String,
    originalSize: json['originalSize'] as int,
    compressedSize: json['compressedSize'] as int,
    compressionRatio: json['compressionRatio'] as double,
    quality: json['quality'] as int,
    format: ImageFormat.values.firstWhere(
      (e) => e.name == json['format'],
      orElse: () => ImageFormat.jpeg,
    ),
  );
}

class ImageInfo {
  final int width;
  final int height;
  final int fileSize;
  final ImageFormat format;
  final bool hasTransparency;

  const ImageInfo({
    required this.width,
    required this.height,
    required this.fileSize,
    required this.format,
    required this.hasTransparency,
  });

  double get aspectRatio => width / height;
  String get resolution => '${width}x$height';
  String get fileSizeFormatted => _formatBytes(fileSize);
  
  static String _formatBytes(int bytes) {
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
}

enum ImageFormat {
  jpeg,
  png,
  webp,
}

class ImageCompressionException extends AppException {
  ImageCompressionException(String message) : super(message);
}