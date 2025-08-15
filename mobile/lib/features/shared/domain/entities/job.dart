import 'package:freezed_annotation/freezed_annotation.dart';

part 'job.freezed.dart';
part 'job.g.dart';

@freezed
class Job with _$Job {
  const factory Job({
    required String id,
    required String customerId,
    required String customerName,
    required String customerEmail,
    String? customerPhone,
    required String serviceId,
    required String serviceName,
    String? serviceDescription,
    required JobStatus status,
    required JobType jobType,
    required JobPriority priority,
    required DateTime scheduledDate,
    DateTime? startedAt,
    DateTime? completedAt,
    DateTime? cancelledAt,
    required JobLocation location,
    String? notes,
    String? customerNotes,
    String? crewNotes,
    @Default([]) List<String> assignedCrewIds,
    @Default([]) List<JobPhoto> photos,
    @Default([]) List<JobTimeEntry> timeEntries,
    JobPricing? pricing,
    double? estimatedDuration,
    double? actualDuration,
    String? equipmentNeeded,
    @Default([]) List<String> tags,
    Map<String, dynamic>? metadata,
    @Default(false) bool requiresSignature,
    String? customerSignature,
    String? crewSignature,
    DateTime? createdAt,
    DateTime? updatedAt,
    String? companyId,
    @Default(false) bool isSynced,
    DateTime? lastSyncAt,
  }) = _Job;

  factory Job.fromJson(Map<String, dynamic> json) => _$JobFromJson(json);
}

@freezed
class JobLocation with _$JobLocation {
  const factory JobLocation({
    required String address,
    String? city,
    String? state,
    String? zipCode,
    String? country,
    double? latitude,
    double? longitude,
    String? specialInstructions,
    String? accessCodes,
    String? contactPerson,
    String? contactPhone,
  }) = _JobLocation;

  factory JobLocation.fromJson(Map<String, dynamic> json) => _$JobLocationFromJson(json);
}

@freezed
class JobPhoto with _$JobPhoto {
  const factory JobPhoto({
    required String id,
    required String jobId,
    required String url,
    String? localPath,
    required PhotoType type,
    String? caption,
    String? description,
    DateTime? takenAt,
    String? takenBy,
    double? latitude,
    double? longitude,
    @Default(false) bool isUploaded,
    DateTime? uploadedAt,
    DateTime? createdAt,
  }) = _JobPhoto;

  factory JobPhoto.fromJson(Map<String, dynamic> json) => _$JobPhotoFromJson(json);
}

@freezed
class JobTimeEntry with _$JobTimeEntry {
  const factory JobTimeEntry({
    required String id,
    required String jobId,
    required String crewMemberId,
    required String crewMemberName,
    required DateTime startTime,
    DateTime? endTime,
    double? breakDuration,
    String? notes,
    double? latitude,
    double? longitude,
    @Default(false) bool isManualEntry,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _JobTimeEntry;

  factory JobTimeEntry.fromJson(Map<String, dynamic> json) => _$JobTimeEntryFromJson(json);
}

@freezed
class JobPricing with _$JobPricing {
  const factory JobPricing({
    required double basePrice,
    double? laborCost,
    double? materialsCost,
    double? equipmentCost,
    double? additionalCosts,
    double? discount,
    double? tax,
    double? totalPrice,
    String? currency,
    String? notes,
  }) = _JobPricing;

  factory JobPricing.fromJson(Map<String, dynamic> json) => _$JobPricingFromJson(json);
}

enum JobStatus {
  @JsonValue('pending')
  pending,
  @JsonValue('scheduled')
  scheduled,
  @JsonValue('in_progress')
  inProgress,
  @JsonValue('paused')
  paused,
  @JsonValue('completed')
  completed,
  @JsonValue('cancelled')
  cancelled,
  @JsonValue('rescheduled')
  rescheduled,
  @JsonValue('requires_follow_up')
  requiresFollowUp,
}

enum JobType {
  @JsonValue('lawn_care')
  lawnCare,
  @JsonValue('landscaping')
  landscaping,
  @JsonValue('tree_service')
  treeService,
  @JsonValue('irrigation')
  irrigation,
  @JsonValue('hardscaping')
  hardscaping,
  @JsonValue('snow_removal')
  snowRemoval,
  @JsonValue('leaf_removal')
  leafRemoval,
  @JsonValue('fertilization')
  fertilization,
  @JsonValue('pest_control')
  pestControl,
  @JsonValue('maintenance')
  maintenance,
  @JsonValue('installation')
  installation,
  @JsonValue('consultation')
  consultation,
  @JsonValue('other')
  other,
}

enum JobPriority {
  @JsonValue('low')
  low,
  @JsonValue('normal')
  normal,
  @JsonValue('high')
  high,
  @JsonValue('urgent')
  urgent,
}

enum PhotoType {
  @JsonValue('before')
  before,
  @JsonValue('after')
  after,
  @JsonValue('progress')
  progress,
  @JsonValue('issue')
  issue,
  @JsonValue('equipment')
  equipment,
  @JsonValue('materials')
  materials,
  @JsonValue('other')
  other,
}

// Extension methods for enum display
extension JobStatusExtension on JobStatus {
  String get displayName {
    switch (this) {
      case JobStatus.pending:
        return 'Pending';
      case JobStatus.scheduled:
        return 'Scheduled';
      case JobStatus.inProgress:
        return 'In Progress';
      case JobStatus.paused:
        return 'Paused';
      case JobStatus.completed:
        return 'Completed';
      case JobStatus.cancelled:
        return 'Cancelled';
      case JobStatus.rescheduled:
        return 'Rescheduled';
      case JobStatus.requiresFollowUp:
        return 'Requires Follow-up';
    }
  }

  String get description {
    switch (this) {
      case JobStatus.pending:
        return 'Job is awaiting confirmation';
      case JobStatus.scheduled:
        return 'Job is scheduled and ready to start';
      case JobStatus.inProgress:
        return 'Work is currently being performed';
      case JobStatus.paused:
        return 'Work has been temporarily paused';
      case JobStatus.completed:
        return 'Job has been successfully completed';
      case JobStatus.cancelled:
        return 'Job has been cancelled';
      case JobStatus.rescheduled:
        return 'Job has been rescheduled to a new date';
      case JobStatus.requiresFollowUp:
        return 'Job requires additional follow-up work';
    }
  }
}

extension JobTypeExtension on JobType {
  String get displayName {
    switch (this) {
      case JobType.lawnCare:
        return 'Lawn Care';
      case JobType.landscaping:
        return 'Landscaping';
      case JobType.treeService:
        return 'Tree Service';
      case JobType.irrigation:
        return 'Irrigation';
      case JobType.hardscaping:
        return 'Hardscaping';
      case JobType.snowRemoval:
        return 'Snow Removal';
      case JobType.leafRemoval:
        return 'Leaf Removal';
      case JobType.fertilization:
        return 'Fertilization';
      case JobType.pestControl:
        return 'Pest Control';
      case JobType.maintenance:
        return 'Maintenance';
      case JobType.installation:
        return 'Installation';
      case JobType.consultation:
        return 'Consultation';
      case JobType.other:
        return 'Other';
    }
  }
}

extension JobPriorityExtension on JobPriority {
  String get displayName {
    switch (this) {
      case JobPriority.low:
        return 'Low';
      case JobPriority.normal:
        return 'Normal';
      case JobPriority.high:
        return 'High';
      case JobPriority.urgent:
        return 'Urgent';
    }
  }
}

extension PhotoTypeExtension on PhotoType {
  String get displayName {
    switch (this) {
      case PhotoType.before:
        return 'Before';
      case PhotoType.after:
        return 'After';
      case PhotoType.progress:
        return 'Progress';
      case PhotoType.issue:
        return 'Issue';
      case PhotoType.equipment:
        return 'Equipment';
      case PhotoType.materials:
        return 'Materials';
      case PhotoType.other:
        return 'Other';
    }
  }
}