import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:reactive_forms/reactive_forms.dart';
import '../../../../core/config/app_config.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/constants/app_constants.dart';
import '../../../../shared/widgets/custom_text_field.dart';
import '../../../../shared/widgets/custom_button.dart';
import '../providers/auth_provider.dart';

class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  late FormGroup form;
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;

  @override
  void initState() {
    super.initState();
    _initForm();
  }

  void _initForm() {
    form = FormGroup({
      'firstName': FormControl<String>(
        validators: [
          Validators.required,
          Validators.maxLength(AppConstants.maxNameLength),
        ],
      ),
      'lastName': FormControl<String>(
        validators: [
          Validators.required,
          Validators.maxLength(AppConstants.maxNameLength),
        ],
      ),
      'email': FormControl<String>(
        validators: [
          Validators.required,
          Validators.email,
        ],
      ),
      'phone': FormControl<String>(
        validators: [],
      ),
      'userType': FormControl<String>(
        value: 'customer',
        validators: [
          Validators.required,
        ],
      ),
      'companyName': FormControl<String>(
        validators: [],
      ),
      'password': FormControl<String>(
        validators: [
          Validators.required,
          Validators.minLength(AppConstants.minPasswordLength),
        ],
      ),
      'confirmPassword': FormControl<String>(
        validators: [
          Validators.required,
        ],
      ),
    });
  }

  Future<void> _handleRegister() async {
    if (form.invalid) {
      form.markAllAsTouched();
      return;
    }
    
    final firstName = form.control('firstName').value as String;
    final lastName = form.control('lastName').value as String;
    final email = form.control('email').value as String;
    final phone = form.control('phone').value as String?;
    final userType = form.control('userType').value as String;
    final companyName = form.control('companyName').value as String?;
    final password = form.control('password').value as String;
    final confirmPassword = form.control('confirmPassword').value as String;
    
    if (password != confirmPassword) {
      // Show error for password mismatch
      return;
    }

    final success = await ref.read(authProvider.notifier).register(
      email: email,
      password: password,
      firstName: firstName,
      lastName: lastName,
      userType: userType,
      phone: phone?.isNotEmpty == true ? phone : null,
      companyName: companyName?.isNotEmpty == true ? companyName : null,
    );

    if (success && mounted) {
      final user = ref.read(currentUserProvider);
      if (user != null) {
        if (user.userType == 'customer') {
          context.go('/customer');
        } else if (user.userType == 'crew') {
          context.go('/crew');
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // final authState = ref.watch(authProvider);
    final isLoading = ref.watch(isLoadingProvider);
    final authError = ref.watch(authErrorProvider);

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24.0),
          child: ReactiveForm(
            formGroup: form,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 20),
                
                // Logo
                Center(
                  child: Container(
                    width: 80,
                    height: 80,
                    decoration: const BoxDecoration(
                      color: AppColors.primary,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.grass,
                      size: 40,
                      color: Colors.white,
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                
                // Welcome Text
                Text(
                  'Create Account',
                  style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'Join ${AppConfig.appName} today',
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
                
                // Error Message
                if (authError != null)
                  Container(
                    padding: const EdgeInsets.all(12),
                    margin: const EdgeInsets.only(bottom: 16),
                    decoration: BoxDecoration(
                      color: AppColors.errorLight,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: AppColors.error),
                    ),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.error_outline,
                          color: AppColors.error,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            authError,
                            style: const TextStyle(color: AppColors.error),
                          ),
                        ),
                        IconButton(
                          onPressed: () => ref.read(authProvider.notifier).clearError(),
                          icon: const Icon(
                            Icons.close,
                            color: AppColors.error,
                            size: 20,
                          ),
                        ),
                      ],
                    ),
                  ),
                
                // Name Fields
                Row(
                  children: [
                    Expanded(
                      child: CustomTextField(
                        formControlName: 'firstName',
                        labelText: 'First Name',
                        hintText: 'Enter first name',
                        prefixIcon: Icons.person_outlined,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: CustomTextField(
                        formControlName: 'lastName',
                        labelText: 'Last Name',
                        hintText: 'Enter last name',
                        prefixIcon: Icons.person_outlined,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                
                // Email Field
                CustomTextField(
                  formControlName: 'email',
                  labelText: 'Email',
                  hintText: 'Enter your email',
                  keyboardType: TextInputType.emailAddress,
                  prefixIcon: Icons.email_outlined,
                ),
                const SizedBox(height: 16),
                
                // Phone Field
                CustomTextField(
                  formControlName: 'phone',
                  labelText: 'Phone (Optional)',
                  hintText: 'Enter your phone number',
                  keyboardType: TextInputType.phone,
                  prefixIcon: Icons.phone_outlined,
                ),
                const SizedBox(height: 16),
                
                // User Type Field
                CustomDropdownField<String>(
                  formControlName: 'userType',
                  labelText: 'Account Type',
                  prefixIcon: Icons.work_outlined,
                  items: const [
                    DropdownMenuItem(
                      value: 'customer',
                      child: Text('Customer'),
                    ),
                    DropdownMenuItem(
                      value: 'crew',
                      child: Text('Crew Member'),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                
                // Company Name (conditional)
                ReactiveValueListenableBuilder<String>(
                  formControlName: 'userType',
                  builder: (context, control, child) {
                    if (control.value == 'crew') {
                      return Column(
                        children: [
                          CustomTextField(
                            formControlName: 'companyName',
                            labelText: 'Company Name (Optional)',
                            hintText: 'Enter company name',
                            prefixIcon: Icons.business_outlined,
                          ),
                          const SizedBox(height: 16),
                        ],
                      );
                    }
                    return const SizedBox.shrink();
                  },
                ),
                
                // Password Field
                CustomTextField(
                  formControlName: 'password',
                  labelText: 'Password',
                  hintText: 'Enter your password',
                  obscureText: _obscurePassword,
                  prefixIcon: Icons.lock_outlined,
                  suffixIcon: IconButton(
                    onPressed: () {
                      setState(() {
                        _obscurePassword = !_obscurePassword;
                      });
                    },
                    icon: Icon(
                      _obscurePassword ? Icons.visibility : Icons.visibility_off,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                
                // Confirm Password Field
                CustomTextField(
                  formControlName: 'confirmPassword',
                  labelText: 'Confirm Password',
                  hintText: 'Confirm your password',
                  obscureText: _obscureConfirmPassword,
                  prefixIcon: Icons.lock_outlined,
                  suffixIcon: IconButton(
                    onPressed: () {
                      setState(() {
                        _obscureConfirmPassword = !_obscureConfirmPassword;
                      });
                    },
                    icon: Icon(
                      _obscureConfirmPassword ? Icons.visibility : Icons.visibility_off,
                    ),
                  ),
                ),
                
                const SizedBox(height: 8),
                const SizedBox(height: 24),
                
                // Register Button
                CustomButton(
                  text: 'Create Account',
                  onPressed: isLoading ? null : _handleRegister,
                  isLoading: isLoading,
                ),
                const SizedBox(height: 24),
                
                // Login Link
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Already have an account? ',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    TextButton(
                      onPressed: () => context.pop(),
                      child: const Text('Sign In'),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }
}