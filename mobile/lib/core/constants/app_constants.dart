class AppConstants {
  // Storage Keys
  static const String keyAccessToken = 'access_token';
  static const String keyRefreshToken = 'refresh_token';
  static const String keyUserId = 'user_id';
  static const String keyUserType = 'user_type';
  static const String keyBiometricEnabled = 'biometric_enabled';
  static const String keyThemeMode = 'theme_mode';
  static const String keyLanguage = 'language';
  static const String keyOnboardingCompleted = 'onboarding_completed';
  static const String keyLastSyncTime = 'last_sync_time';
  static const String keyOfflineData = 'offline_data';

  // User Types
  static const String userTypeCustomer = 'customer';
  static const String userTypeCrew = 'crew';
  static const String userTypeAdmin = 'admin';

  // Job Status
  static const String jobStatusPending = 'pending';
  static const String jobStatusScheduled = 'scheduled';
  static const String jobStatusInProgress = 'in_progress';
  static const String jobStatusCompleted = 'completed';
  static const String jobStatusCancelled = 'cancelled';

  // Quote Status
  static const String quoteStatusDraft = 'draft';
  static const String quoteStatusSent = 'sent';
  static const String quoteStatusApproved = 'approved';
  static const String quoteStatusRejected = 'rejected';
  static const String quoteStatusExpired = 'expired';

  // Invoice Status
  static const String invoiceStatusDraft = 'draft';
  static const String invoiceStatusSent = 'sent';
  static const String invoiceStatusPaid = 'paid';
  static const String invoiceStatusOverdue = 'overdue';
  static const String invoiceStatusCancelled = 'cancelled';

  // Notification Types
  static const String notificationTypeJobUpdate = 'job_update';
  static const String notificationTypeQuoteUpdate = 'quote_update';
  static const String notificationTypeInvoiceUpdate = 'invoice_update';
  static const String notificationTypeMessage = 'message';
  static const String notificationTypeReminder = 'reminder';

  // API Endpoints
  static const String endpointAuth = '/auth';
  static const String endpointLogin = '/auth/login';
  static const String endpointRegister = '/auth/register';
  static const String endpointRefreshToken = '/auth/refresh';
  static const String endpointLogout = '/auth/logout';
  static const String endpointProfile = '/profile';
  static const String endpointJobs = '/jobs';
  static const String endpointQuotes = '/quotes';
  static const String endpointInvoices = '/invoices';
  static const String endpointCustomers = '/customers';
  static const String endpointProperties = '/properties';
  static const String endpointEquipment = '/equipment';
  static const String endpointNotifications = '/notifications';
  static const String endpointUpload = '/upload';

  // Error Messages
  static const String errorGeneral = 'An error occurred. Please try again.';
  static const String errorNetwork = 'Network error. Please check your connection.';
  static const String errorAuth = 'Authentication failed. Please login again.';
  static const String errorPermission = 'Permission denied.';
  static const String errorNotFound = 'Resource not found.';
  static const String errorValidation = 'Validation error. Please check your input.';
  static const String errorServer = 'Server error. Please try again later.';

  // Animation Durations
  static const Duration animationFast = Duration(milliseconds: 200);
  static const Duration animationNormal = Duration(milliseconds: 300);
  static const Duration animationSlow = Duration(milliseconds: 500);

  // UI Constants
  static const double paddingSmall = 8.0;
  static const double paddingMedium = 16.0;
  static const double paddingLarge = 24.0;
  static const double paddingXLarge = 32.0;

  static const double radiusSmall = 4.0;
  static const double radiusMedium = 8.0;
  static const double radiusLarge = 16.0;
  static const double radiusXLarge = 24.0;

  static const double elevationLow = 2.0;
  static const double elevationMedium = 4.0;
  static const double elevationHigh = 8.0;

  // Date Formats
  static const String dateFormatDisplay = 'MMM dd, yyyy';
  static const String dateFormatApi = 'yyyy-MM-dd';
  static const String dateTimeFormatDisplay = 'MMM dd, yyyy HH:mm';
  static const String dateTimeFormatApi = 'yyyy-MM-ddTHH:mm:ssZ';
  static const String timeFormatDisplay = 'HH:mm';

  // Validation
  static const int minPasswordLength = 8;
  static const int maxNameLength = 50;
  static const int maxDescriptionLength = 500;
  static const int maxAddressLength = 200;

  // Retry Configuration
  static const int maxRetryAttempts = 3;
  static const Duration retryDelay = Duration(seconds: 2);

  // Sync Configuration
  static const Duration syncInterval = Duration(minutes: 15);
  static const Duration backgroundSyncInterval = Duration(hours: 1);
}