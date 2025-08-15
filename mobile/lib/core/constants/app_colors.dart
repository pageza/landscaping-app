import 'package:flutter/material.dart';

class AppColors {
  // Primary Colors
  static const Color primary = Color(0xFF2E7D32); // Green for landscaping
  static const Color primaryDark = Color(0xFF1B5E20);
  static const Color primaryLight = Color(0xFF4CAF50);
  static const Color primaryVariant = Color(0xFF388E3C);

  // Secondary Colors
  static const Color secondary = Color(0xFF795548); // Brown for earth/soil
  static const Color secondaryDark = Color(0xFF5D4037);
  static const Color secondaryLight = Color(0xFF8D6E63);
  static const Color secondaryVariant = Color(0xFF6D4C41);

  // Accent Colors
  static const Color accent = Color(0xFFFFB74D); // Orange for equipment/tools
  static const Color accentDark = Color(0xFFFF9800);
  static const Color accentLight = Color(0xFFFFCC02);

  // Neutral Colors
  static const Color background = Color(0xFFFAFAFA);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFF5F5F5);
  
  // Text Colors
  static const Color textPrimary = Color(0xFF212121);
  static const Color textSecondary = Color(0xFF757575);
  static const Color textTertiary = Color(0xFF9E9E9E);
  static const Color textOnPrimary = Color(0xFFFFFFFF);
  static const Color textOnSecondary = Color(0xFFFFFFFF);
  static const Color textOnSurface = Color(0xFF212121);

  // Status Colors
  static const Color success = Color(0xFF4CAF50);
  static const Color successLight = Color(0xFFC8E6C9);
  static const Color warning = Color(0xFFFF9800);
  static const Color warningLight = Color(0xFFFFE0B2);
  static const Color error = Color(0xFFF44336);
  static const Color errorLight = Color(0xFFFFCDD2);
  static const Color info = Color(0xFF2196F3);
  static const Color infoLight = Color(0xFFBBDEFB);

  // Job Status Colors
  static const Color jobPending = Color(0xFFFF9800);
  static const Color jobScheduled = Color(0xFF2196F3);
  static const Color jobInProgress = Color(0xFF9C27B0);
  static const Color jobCompleted = Color(0xFF4CAF50);
  static const Color jobCancelled = Color(0xFFF44336);

  // Quote Status Colors
  static const Color quoteDraft = Color(0xFF9E9E9E);
  static const Color quoteSent = Color(0xFF2196F3);
  static const Color quoteApproved = Color(0xFF4CAF50);
  static const Color quoteRejected = Color(0xFFF44336);
  static const Color quoteExpired = Color(0xFF795548);

  // Invoice Status Colors
  static const Color invoiceDraft = Color(0xFF9E9E9E);
  static const Color invoiceSent = Color(0xFF2196F3);
  static const Color invoicePaid = Color(0xFF4CAF50);
  static const Color invoiceOverdue = Color(0xFFF44336);
  static const Color invoiceCancelled = Color(0xFF795548);

  // Border Colors
  static const Color border = Color(0xFFE0E0E0);
  static const Color borderDark = Color(0xFFBDBDBD);
  static const Color borderLight = Color(0xFFF5F5F5);

  // Divider Colors
  static const Color divider = Color(0xFFE0E0E0);
  static const Color dividerLight = Color(0xFFF5F5F5);

  // Shadow Colors
  static const Color shadow = Color(0x1F000000);
  static const Color shadowLight = Color(0x0F000000);
  static const Color shadowDark = Color(0x3F000000);

  // Overlay Colors
  static const Color overlay = Color(0x80000000);
  static const Color overlayLight = Color(0x40000000);
  static const Color overlayDark = Color(0xB3000000);

  // Dark Theme Colors
  static const Color backgroundDark = Color(0xFF121212);
  static const Color surfaceDark = Color(0xFF1E1E1E);
  static const Color surfaceVariantDark = Color(0xFF2D2D2D);
  static const Color textPrimaryDark = Color(0xFFFFFFFF);
  static const Color textSecondaryDark = Color(0xFFB3B3B3);
  static const Color textTertiaryDark = Color(0xFF8C8C8C);

  // Equipment Categories
  static const Color equipmentMowing = Color(0xFF4CAF50);
  static const Color equipmentTrimming = Color(0xFF2196F3);
  static const Color equipmentBlowing = Color(0xFFFF9800);
  static const Color equipmentWatering = Color(0xFF00BCD4);
  static const Color equipmentPlanting = Color(0xFF8BC34A);
  static const Color equipmentLandscaping = Color(0xFF795548);

  // Time-based Colors
  static const Color morning = Color(0xFFFFEB3B);
  static const Color afternoon = Color(0xFFFF9800);
  static const Color evening = Color(0xFF673AB7);
  static const Color night = Color(0xFF3F51B5);

  // Weather Colors
  static const Color sunny = Color(0xFFFFEB3B);
  static const Color cloudy = Color(0xFF9E9E9E);
  static const Color rainy = Color(0xFF2196F3);
  static const Color snowy = Color(0xFFE1F5FE);
  static const Color stormy = Color(0xFF37474F);

  // Helper method to get color by status
  static Color getJobStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return jobPending;
      case 'scheduled':
        return jobScheduled;
      case 'in_progress':
        return jobInProgress;
      case 'completed':
        return jobCompleted;
      case 'cancelled':
        return jobCancelled;
      default:
        return textSecondary;
    }
  }

  static Color getQuoteStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return quoteDraft;
      case 'sent':
        return quoteSent;
      case 'approved':
        return quoteApproved;
      case 'rejected':
        return quoteRejected;
      case 'expired':
        return quoteExpired;
      default:
        return textSecondary;
    }
  }

  static Color getInvoiceStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return invoiceDraft;
      case 'sent':
        return invoiceSent;
      case 'paid':
        return invoicePaid;
      case 'overdue':
        return invoiceOverdue;
      case 'cancelled':
        return invoiceCancelled;
      default:
        return textSecondary;
    }
  }
}