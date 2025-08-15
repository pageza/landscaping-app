import 'package:freezed_annotation/freezed_annotation.dart';
import 'job.dart';

part 'service.freezed.dart';
part 'service.g.dart';

@freezed
class Service with _$Service {
  const factory Service({
    required String id,
    required String name,
    required String description,
    required ServiceCategory category,
    required JobType jobType,
    required double basePrice,
    String? priceUnit,
    double? estimatedDuration,
    String? durationUnit,
    @Default([]) List<String> requiredEquipment,
    @Default([]) List<String> optionalEquipment,
    @Default([]) List<String> skillsRequired,
    String? seasonality,
    @Default(true) bool isActive,
    String? imageUrl,
    @Default([]) List<String> tags,
    Map<String, dynamic>? metadata,
    String? companyId,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _Service;

  factory Service.fromJson(Map<String, dynamic> json) => _$ServiceFromJson(json);
}

@freezed
class ServiceCategory with _$ServiceCategory {
  const factory ServiceCategory({
    required String id,
    required String name,
    required String description,
    String? iconName,
    String? color,
    @Default(0) int sortOrder,
    @Default(true) bool isActive,
    String? parentCategoryId,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _ServiceCategory;

  factory ServiceCategory.fromJson(Map<String, dynamic> json) => _$ServiceCategoryFromJson(json);
}

@freezed
class ServicePricing with _$ServicePricing {
  const factory ServicePricing({
    required String serviceId,
    required PricingType type,
    required double basePrice,
    double? minPrice,
    double? maxPrice,
    String? unit,
    @Default([]) List<PricingTier> tiers,
    @Default([]) List<PricingModifier> modifiers,
    String? description,
    @Default(true) bool isActive,
    DateTime? effectiveFrom,
    DateTime? effectiveTo,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _ServicePricing;

  factory ServicePricing.fromJson(Map<String, dynamic> json) => _$ServicePricingFromJson(json);
}

@freezed
class PricingTier with _$PricingTier {
  const factory PricingTier({
    required String name,
    required double minQuantity,
    double? maxQuantity,
    required double price,
    String? description,
  }) = _PricingTier;

  factory PricingTier.fromJson(Map<String, dynamic> json) => _$PricingTierFromJson(json);
}

@freezed
class PricingModifier with _$PricingModifier {
  const factory PricingModifier({
    required String name,
    required ModifierType type,
    required double value,
    String? condition,
    String? description,
    @Default(true) bool isActive,
  }) = _PricingModifier;

  factory PricingModifier.fromJson(Map<String, dynamic> json) => _$PricingModifierFromJson(json);
}

enum PricingType {
  @JsonValue('fixed')
  fixed,
  @JsonValue('hourly')
  hourly,
  @JsonValue('per_unit')
  perUnit,
  @JsonValue('per_sqft')
  perSquareFoot,
  @JsonValue('per_linear_ft')
  perLinearFoot,
  @JsonValue('tiered')
  tiered,
  @JsonValue('custom')
  custom,
}

enum ModifierType {
  @JsonValue('percentage')
  percentage,
  @JsonValue('fixed_amount')
  fixedAmount,
  @JsonValue('multiplier')
  multiplier,
}

// Extension methods
extension PricingTypeExtension on PricingType {
  String get displayName {
    switch (this) {
      case PricingType.fixed:
        return 'Fixed Price';
      case PricingType.hourly:
        return 'Hourly Rate';
      case PricingType.perUnit:
        return 'Per Unit';
      case PricingType.perSquareFoot:
        return 'Per Square Foot';
      case PricingType.perLinearFoot:
        return 'Per Linear Foot';
      case PricingType.tiered:
        return 'Tiered Pricing';
      case PricingType.custom:
        return 'Custom Pricing';
    }
  }
}

extension ModifierTypeExtension on ModifierType {
  String get displayName {
    switch (this) {
      case ModifierType.percentage:
        return 'Percentage';
      case ModifierType.fixedAmount:
        return 'Fixed Amount';
      case ModifierType.multiplier:
        return 'Multiplier';
    }
  }
}