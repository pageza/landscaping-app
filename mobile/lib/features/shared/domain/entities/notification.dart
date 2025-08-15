import 'package:freezed_annotation/freezed_annotation.dart';

part 'notification.freezed.dart';
part 'notification.g.dart';

@freezed
class AppNotification with _$AppNotification {
  const factory AppNotification({
    required String id,
    required String title,
    required String message,
    required NotificationType type,
    required NotificationPriority priority,
    required String userId,
    String? jobId,
    String? relatedEntityId,
    String? relatedEntityType,
    Map<String, dynamic>? data,
    String? imageUrl,
    String? actionUrl,
    @Default([]) List<NotificationAction> actions,
    @Default(false) bool isRead,
    @Default(false) bool isArchived,
    DateTime? readAt,
    DateTime? scheduledFor,
    DateTime? expiresAt,
    String? channel,
    @Default(true) bool isDelivered,
    DateTime? deliveredAt,
    String? deviceToken,
    String? platformType,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _AppNotification;

  factory AppNotification.fromJson(Map<String, dynamic> json) => _$AppNotificationFromJson(json);
}

@freezed
class NotificationAction with _$NotificationAction {
  const factory NotificationAction({
    required String id,
    required String title,
    required String action,
    String? url,
    Map<String, dynamic>? data,
    @Default(false) bool destructive,
  }) = _NotificationAction;

  factory NotificationAction.fromJson(Map<String, dynamic> json) => _$NotificationActionFromJson(json);
}

@freezed
class NotificationSettings with _$NotificationSettings {
  const factory NotificationSettings({
    required String userId,
    @Default(true) bool jobUpdates,
    @Default(true) bool scheduleChanges,
    @Default(true) bool newJobs,
    @Default(true) bool paymentReminders,
    @Default(true) bool weatherAlerts,
    @Default(true) bool equipmentMaintenance,
    @Default(true) bool teamUpdates,
    @Default(true) bool systemNotifications,
    @Default(true) bool pushNotifications,
    @Default(true) bool emailNotifications,
    @Default(false) bool smsNotifications,
    @Default('immediately') String frequency,
    String? quietHoursStart,
    String? quietHoursEnd,
    @Default([]) List<String> mutedChannels,
    DateTime? updatedAt,
  }) = _NotificationSettings;

  factory NotificationSettings.fromJson(Map<String, dynamic> json) => _$NotificationSettingsFromJson(json);
}

enum NotificationType {
  @JsonValue('job_update')
  jobUpdate,
  @JsonValue('job_assigned')
  jobAssigned,
  @JsonValue('job_completed')
  jobCompleted,
  @JsonValue('job_cancelled')
  jobCancelled,
  @JsonValue('schedule_change')
  scheduleChange,
  @JsonValue('payment_reminder')
  paymentReminder,
  @JsonValue('payment_received')
  paymentReceived,
  @JsonValue('weather_alert')
  weatherAlert,
  @JsonValue('equipment_maintenance')
  equipmentMaintenance,
  @JsonValue('team_message')
  teamMessage,
  @JsonValue('system_update')
  systemUpdate,
  @JsonValue('promotional')
  promotional,
  @JsonValue('reminder')
  reminder,
  @JsonValue('alert')
  alert,
  @JsonValue('info')
  info,
}

enum NotificationPriority {
  @JsonValue('low')
  low,
  @JsonValue('normal')
  normal,
  @JsonValue('high')
  high,
  @JsonValue('urgent')
  urgent,
}

// Extension methods
extension NotificationTypeExtension on NotificationType {
  String get displayName {
    switch (this) {
      case NotificationType.jobUpdate:
        return 'Job Update';
      case NotificationType.jobAssigned:
        return 'Job Assigned';
      case NotificationType.jobCompleted:
        return 'Job Completed';
      case NotificationType.jobCancelled:
        return 'Job Cancelled';
      case NotificationType.scheduleChange:
        return 'Schedule Change';
      case NotificationType.paymentReminder:
        return 'Payment Reminder';
      case NotificationType.paymentReceived:
        return 'Payment Received';
      case NotificationType.weatherAlert:
        return 'Weather Alert';
      case NotificationType.equipmentMaintenance:
        return 'Equipment Maintenance';
      case NotificationType.teamMessage:
        return 'Team Message';
      case NotificationType.systemUpdate:
        return 'System Update';
      case NotificationType.promotional:
        return 'Promotional';
      case NotificationType.reminder:
        return 'Reminder';
      case NotificationType.alert:
        return 'Alert';
      case NotificationType.info:
        return 'Information';
    }
  }

  String get description {
    switch (this) {
      case NotificationType.jobUpdate:
        return 'Updates about your jobs';
      case NotificationType.jobAssigned:
        return 'New job assignments';
      case NotificationType.jobCompleted:
        return 'Job completion notifications';
      case NotificationType.jobCancelled:
        return 'Job cancellation alerts';
      case NotificationType.scheduleChange:
        return 'Schedule modifications';
      case NotificationType.paymentReminder:
        return 'Payment due reminders';
      case NotificationType.paymentReceived:
        return 'Payment confirmations';
      case NotificationType.weatherAlert:
        return 'Weather-related alerts';
      case NotificationType.equipmentMaintenance:
        return 'Equipment maintenance reminders';
      case NotificationType.teamMessage:
        return 'Messages from team members';
      case NotificationType.systemUpdate:
        return 'System and app updates';
      case NotificationType.promotional:
        return 'Promotional offers and news';
      case NotificationType.reminder:
        return 'General reminders';
      case NotificationType.alert:
        return 'Important alerts';
      case NotificationType.info:
        return 'General information';
    }
  }
}

extension NotificationPriorityExtension on NotificationPriority {
  String get displayName {
    switch (this) {
      case NotificationPriority.low:
        return 'Low';
      case NotificationPriority.normal:
        return 'Normal';
      case NotificationPriority.high:
        return 'High';
      case NotificationPriority.urgent:
        return 'Urgent';
    }
  }
}