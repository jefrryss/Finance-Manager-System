import SwiftUI

struct LoginView: View {
    @Binding var isLoggedIn: Bool
    @State private var viewModel = LoginViewModel()
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()
            
            VStack(spacing: 30) {
                VStack(spacing: 12) {
                    Text("С возвращением")
                        .font(.system(size: 32, weight: .bold))
                        .foregroundColor(AppTheme.textPrimary)
                    
                    Text("Войдите в свой аккаунт Finexa")
                        .font(.system(size: 16))
                        .foregroundColor(AppTheme.textSecondary)
                }
                .padding(.top, 40)
                
                VStack(spacing: 16) {
                    FintechTextField(icon: "person.fill", placeholder: "Логин или Email", text: $viewModel.identifier)
                    FintechTextField(icon: "lock.fill", placeholder: "Пароль", text: $viewModel.password, isSecure: true)
                }
                
                if let error = viewModel.errorMessage {
                    Text(error)
                        .foregroundColor(.red)
                        .font(.caption)
                }
                
                FintechButton(
                    title: "Войти",
                    isLoading: viewModel.isLoading,
                    isDisabled: viewModel.identifier.isEmpty || viewModel.password.isEmpty
                ) {
                    Task {
                        if await viewModel.loginUser() {
                            withAnimation {
                                isLoggedIn = true
                            }
                        }
                    }
                }
                
                Button {
                    dismiss()
                } label: {
                    HStack {
                        Text("Нет аккаунта?")
                            .foregroundColor(AppTheme.textSecondary)
                        Text("Создать").bold()
                            .foregroundColor(AppTheme.accent)
                    }
                    .font(.system(size: 14))
                }
                
                Spacer()
            }
            .padding(24)
        }
        .navigationBarBackButtonHidden(true)
    }
}
