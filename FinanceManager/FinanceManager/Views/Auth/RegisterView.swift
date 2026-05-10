import SwiftUI

struct RegisterView: View {
    @Environment(\.dismiss) var dismiss
    @State private var viewModel = RegisterViewModel()
    @Binding var isLoggedIn: Bool

    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground
                
                VStack(spacing: 30) {
                    VStack(spacing: 8) {
                        Text("finexa")
                            .font(.system(size: 32, weight: .black, design: .rounded))
                            .foregroundColor(AppTheme.textPrimary)
                        
                        Text("Регистрация")
                            .font(.subheadline)
                            .foregroundColor(AppTheme.textSecondary)
                    }
                    .padding(.top, 20)
                    
                    VStack(spacing: 20) {
                        FintechTextField(icon: "envelope.fill", placeholder: "Email", text: $viewModel.email)
                            .keyboardType(.emailAddress)
                            .autocapitalization(.none)
                        
                        FintechTextField(icon: "person.fill", placeholder: "Логин", text: $viewModel.login)
                            .autocapitalization(.none)
                        
                        FintechTextField(icon: "lock.fill", placeholder: "Пароль (мин. 6 символов)", text: $viewModel.password, isSecure: true)
                        
                        if let error = viewModel.errorMessage {
                            Text(error)
                                .font(.caption)
                                .foregroundColor(.red)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                    }
                    
                    Spacer()
                    
                    FintechButton(
                        title: "Создать аккаунт",
                        isLoading: viewModel.isLoading,
                        isDisabled: !viewModel.isFormValid
                    ) {
                        Task {
                            if await viewModel.register() {
                                withAnimation(.spring()) {
                                    isLoggedIn = true
                                    dismiss()
                                }
                            }
                        }
                    }
                    .padding(.bottom, 20)
                }
                .padding(.horizontal, 24)
            }
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button { dismiss() } label: {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(AppTheme.textSecondary)
                    }
                }
            }
        }
    }
}
