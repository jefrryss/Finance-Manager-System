import SwiftUI

struct PasswordChangeView: View {
    @Bindable var viewModel: ProfileViewModel
    @State private var newPassword = ""
    @State private var confirmPassword = ""
    @State private var localError: String?
    @State private var isSaving = false
    @Environment(\.dismiss) private var dismiss

    private var passwordsMatch: Bool {
        confirmPassword.isEmpty || newPassword == confirmPassword
    }

    private var canSave: Bool {
        !newPassword.isEmpty && !confirmPassword.isEmpty && passwordsMatch && !isSaving
    }

    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 24) {
                    VStack(spacing: 12) {
                        Text("Сменить пароль")
                            .font(.system(size: 24, weight: .bold))
                            .foregroundColor(AppTheme.textPrimary)

                        Text("Введите новый пароль и подтвердите его")
                            .font(.system(size: 14))
                            .foregroundColor(AppTheme.textSecondary)
                            .multilineTextAlignment(.center)
                    }
                    .padding(.top, 20)

                    VStack(spacing: 16) {
                        FintechTextField(icon: "lock.fill", placeholder: "Новый пароль", text: $newPassword, isSecure: true)
                        FintechTextField(icon: "lock.rotation", placeholder: "Подтвердите пароль", text: $confirmPassword, isSecure: true)
                    }
                    .padding(.horizontal, 24)

                    if !passwordsMatch {
                        Text("Пароли не совпадают")
                            .foregroundColor(.red)
                            .font(.system(size: 14, weight: .medium))
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 24)
                    } else if let error = localError ?? viewModel.errorMessage {
                        Text(error)
                            .foregroundColor(.red)
                            .font(.system(size: 14, weight: .medium))
                            .multilineTextAlignment(.center)
                            .padding(.horizontal, 24)
                    }

                    FintechButton(title: isSaving ? "Сохранение..." : "Сохранить", isLoading: isSaving, isDisabled: !canSave) {
                        Task {
                            localError = nil
                            viewModel.errorMessage = nil

                            guard !newPassword.isEmpty else {
                                localError = "Введите новый пароль"
                                return
                            }

                            guard newPassword == confirmPassword else {
                                localError = "Пароли не совпадают"
                                return
                            }

                            isSaving = true
                            let success = await viewModel.changePassword(newPassword: newPassword)
                            isSaving = false

                            if success {
                                dismiss()
                            } else {
                                localError = viewModel.errorMessage ?? "Не удалось сменить пароль"
                            }
                        }
                    }
                    .padding(.horizontal, 24)

                    Spacer()
                }
                .padding(.bottom, 30)
            }
        }
        .navigationTitle("Пароль")
        .navigationBarTitleDisplayMode(.inline)
    }
}
