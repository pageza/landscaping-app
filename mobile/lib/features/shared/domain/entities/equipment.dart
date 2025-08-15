import 'package:freezed_annotation/freezed_annotation.dart';

part 'equipment.freezed.dart';
part 'equipment.g.dart';

@freezed
class Equipment with _$Equipment {
  const factory Equipment({
    required String id,
    required String name,
    required String model,
    String? brand,
    String? serialNumber,
    required EquipmentType type,
    required EquipmentStatus status,
    String? description,
    DateTime? purchaseDate,
    double? purchasePrice,
    String? supplier,
    String? warrantyInfo,
    DateTime? warrantyExpiry,
    String? location,
    String? assignedTo,
    @Default([]) List<MaintenanceRecord> maintenanceHistory,
    DateTime? lastMaintenanceDate,
    DateTime? nextMaintenanceDate,
    String? qrCode,
    String? imageUrl,
    @Default([]) List<String> tags,
    Map<String, dynamic>? specifications,
    Map<String, dynamic>? metadata,
    String? companyId,
    @Default(true) bool isActive,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _Equipment;

  factory Equipment.fromJson(Map<String, dynamic> json) => _$EquipmentFromJson(json);
}

@freezed
class MaintenanceRecord with _$MaintenanceRecord {
  const factory MaintenanceRecord({
    required String id,
    required String equipmentId,
    required MaintenanceType type,
    required DateTime date,
    required String description,
    String? performedBy,
    double? cost,
    String? notes,
    @Default([]) List<String> partsReplaced,
    DateTime? nextServiceDate,
    String? invoiceNumber,
    String? vendorInfo,
    DateTime? createdAt,
  }) = _MaintenanceRecord;

  factory MaintenanceRecord.fromJson(Map<String, dynamic> json) => _$MaintenanceRecordFromJson(json);
}

@freezed
class EquipmentUsage with _$EquipmentUsage {
  const factory EquipmentUsage({
    required String id,
    required String equipmentId,
    required String jobId,
    required String usedBy,
    required DateTime startTime,
    DateTime? endTime,
    double? hoursUsed,
    double? fuelUsed,
    String? condition,
    String? notes,
    String? issuesReported,
    DateTime? createdAt,
  }) = _EquipmentUsage;

  factory EquipmentUsage.fromJson(Map<String, dynamic> json) => _$EquipmentUsageFromJson(json);
}

enum EquipmentType {
  @JsonValue('mower')
  mower,
  @JsonValue('trimmer')
  trimmer,
  @JsonValue('blower')
  blower,
  @JsonValue('edger')
  edger,
  @JsonValue('chainsaw')
  chainsaw,
  @JsonValue('hedge_trimmer')
  hedgeTrimmer,
  @JsonValue('pressure_washer')
  pressureWasher,
  @JsonValue('spreader')
  spreader,
  @JsonValue('aerator')
  aerator,
  @JsonValue('dethatcher')
  dethatcher,
  @JsonValue('tiller')
  tiller,
  @JsonValue('truck')
  truck,
  @JsonValue('trailer')
  trailer,
  @JsonValue('hand_tools')
  handTools,
  @JsonValue('irrigation_equipment')
  irrigationEquipment,
  @JsonValue('safety_equipment')
  safetyEquipment,
  @JsonValue('other')
  other,
}

enum EquipmentStatus {
  @JsonValue('available')
  available,
  @JsonValue('in_use')
  inUse,
  @JsonValue('maintenance')
  maintenance,
  @JsonValue('repair')
  repair,
  @JsonValue('out_of_service')
  outOfService,
  @JsonValue('retired')
  retired,
}

enum MaintenanceType {
  @JsonValue('routine')
  routine,
  @JsonValue('preventive')
  preventive,
  @JsonValue('corrective')
  corrective,
  @JsonValue('emergency')
  emergency,
  @JsonValue('inspection')
  inspection,
  @JsonValue('cleaning')
  cleaning,
  @JsonValue('calibration')
  calibration,
}

// Extension methods
extension EquipmentTypeExtension on EquipmentType {
  String get displayName {
    switch (this) {
      case EquipmentType.mower:
        return 'Mower';
      case EquipmentType.trimmer:
        return 'Trimmer';
      case EquipmentType.blower:
        return 'Blower';
      case EquipmentType.edger:
        return 'Edger';
      case EquipmentType.chainsaw:
        return 'Chainsaw';
      case EquipmentType.hedgeTrimmer:
        return 'Hedge Trimmer';
      case EquipmentType.pressureWasher:
        return 'Pressure Washer';
      case EquipmentType.spreader:
        return 'Spreader';
      case EquipmentType.aerator:
        return 'Aerator';
      case EquipmentType.dethatcher:
        return 'Dethatcher';
      case EquipmentType.tiller:
        return 'Tiller';
      case EquipmentType.truck:
        return 'Truck';
      case EquipmentType.trailer:
        return 'Trailer';
      case EquipmentType.handTools:
        return 'Hand Tools';
      case EquipmentType.irrigationEquipment:
        return 'Irrigation Equipment';
      case EquipmentType.safetyEquipment:
        return 'Safety Equipment';
      case EquipmentType.other:
        return 'Other';
    }
  }
}

extension EquipmentStatusExtension on EquipmentStatus {
  String get displayName {
    switch (this) {
      case EquipmentStatus.available:
        return 'Available';
      case EquipmentStatus.inUse:
        return 'In Use';
      case EquipmentStatus.maintenance:
        return 'Maintenance';
      case EquipmentStatus.repair:
        return 'Repair';
      case EquipmentStatus.outOfService:
        return 'Out of Service';
      case EquipmentStatus.retired:
        return 'Retired';
    }
  }
}

extension MaintenanceTypeExtension on MaintenanceType {
  String get displayName {
    switch (this) {
      case MaintenanceType.routine:
        return 'Routine';
      case MaintenanceType.preventive:
        return 'Preventive';
      case MaintenanceType.corrective:
        return 'Corrective';
      case MaintenanceType.emergency:
        return 'Emergency';
      case MaintenanceType.inspection:
        return 'Inspection';
      case MaintenanceType.cleaning:
        return 'Cleaning';
      case MaintenanceType.calibration:
        return 'Calibration';
    }
  }
}