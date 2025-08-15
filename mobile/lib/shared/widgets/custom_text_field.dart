import 'package:flutter/material.dart';
import 'package:reactive_forms/reactive_forms.dart';
import '../../core/constants/app_colors.dart';

class CustomTextField extends StatelessWidget {
  final String formControlName;
  final String labelText;
  final String? hintText;
  final IconData? prefixIcon;
  final Widget? suffixIcon;
  final TextInputType keyboardType;
  final bool obscureText;
  final int maxLines;
  final bool enabled;
  final String? helperText;
  final List<String? Function(AbstractControl<dynamic>)>? validators;

  const CustomTextField({
    super.key,
    required this.formControlName,
    required this.labelText,
    this.hintText,
    this.prefixIcon,
    this.suffixIcon,
    this.keyboardType = TextInputType.text,
    this.obscureText = false,
    this.maxLines = 1,
    this.enabled = true,
    this.helperText,
    this.validators,
  });

  @override
  Widget build(BuildContext context) {
    return ReactiveTextField<String>(
      formControlName: formControlName,
      keyboardType: keyboardType,
      obscureText: obscureText,
      maxLines: maxLines,
      decoration: InputDecoration(
        labelText: labelText,
        hintText: hintText,
        helperText: helperText,
        prefixIcon: prefixIcon != null ? Icon(prefixIcon) : null,
        suffixIcon: suffixIcon,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.primary, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error, width: 2),
        ),
        filled: true,
        fillColor: enabled 
            ? Theme.of(context).colorScheme.surface 
            : AppColors.surfaceVariant,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 16,
        ),
      ),
      validationMessages: {
        ValidationMessage.required: (_) => '$labelText is required',
        ValidationMessage.email: (_) => 'Please enter a valid email address',
        ValidationMessage.minLength: (error) => 
            '$labelText must be at least ${(error as Map)['requiredLength']} characters',
        ValidationMessage.maxLength: (error) => 
            '$labelText must not exceed ${(error as Map)['requiredLength']} characters',
      },
    );
  }
}

class CustomDropdownField<T> extends StatelessWidget {
  final String formControlName;
  final String labelText;
  final String? hintText;
  final IconData? prefixIcon;
  final List<DropdownMenuItem<T>> items;
  final bool enabled;
  final String? helperText;

  const CustomDropdownField({
    super.key,
    required this.formControlName,
    required this.labelText,
    required this.items,
    this.hintText,
    this.prefixIcon,
    this.enabled = true,
    this.helperText,
  });

  @override
  Widget build(BuildContext context) {
    return ReactiveDropdownField<T>(
      formControlName: formControlName,
      items: items,
      decoration: InputDecoration(
        labelText: labelText,
        hintText: hintText,
        helperText: helperText,
        prefixIcon: prefixIcon != null ? Icon(prefixIcon) : null,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.primary, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error, width: 2),
        ),
        filled: true,
        fillColor: enabled 
            ? Theme.of(context).colorScheme.surface 
            : AppColors.surfaceVariant,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 16,
        ),
      ),
      validationMessages: {
        ValidationMessage.required: (_) => '$labelText is required',
      },
    );
  }
}

class CustomTextArea extends StatelessWidget {
  final String formControlName;
  final String labelText;
  final String? hintText;
  final int maxLines;
  final int? maxLength;
  final bool enabled;
  final String? helperText;

  const CustomTextArea({
    super.key,
    required this.formControlName,
    required this.labelText,
    this.hintText,
    this.maxLines = 4,
    this.maxLength,
    this.enabled = true,
    this.helperText,
  });

  @override
  Widget build(BuildContext context) {
    return ReactiveTextField<String>(
      formControlName: formControlName,
      keyboardType: TextInputType.multiline,
      maxLines: maxLines,
      maxLength: maxLength,
      decoration: InputDecoration(
        labelText: labelText,
        hintText: hintText,
        helperText: helperText,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.primary, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.error, width: 2),
        ),
        filled: true,
        fillColor: enabled 
            ? Theme.of(context).colorScheme.surface 
            : AppColors.surfaceVariant,
        contentPadding: const EdgeInsets.all(16),
        alignLabelWithHint: true,
      ),
      validationMessages: {
        ValidationMessage.required: (_) => '$labelText is required',
        ValidationMessage.maxLength: (error) => 
            '$labelText must not exceed ${(error as Map)['requiredLength']} characters',
      },
    );
  }
}