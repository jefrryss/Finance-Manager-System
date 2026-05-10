import SwiftUI

struct WelcomeView: View {
    @Binding var isLoggedIn: Bool
    @State private var showingRegister = false
    
    var body: some View {
        // Оборачиваем в NavigationStack для работы переходов (NavigationLink)
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground.ignoresSafeArea()
                
                VStack(spacing: 0) {
                    Spacer()
                    
                    VStack(spacing: 20) {
                        ZStack {
                            Circle()
                                .fill(AppTheme.accent.opacity(0.1))
                                .frame(width: 100, height: 100)
                            
                            Image(systemName: "chart.line.uptrend.xyaxis")
                                .font(.system(size: 44, weight: .semibold))
                                .foregroundColor(AppTheme.accent)
                        }
                        
                        Text("finexa")
                            .font(.system(size: 48, weight: .black, design: .rounded))
                            .foregroundColor(AppTheme.textPrimary)
                        
                        Text("трать с умом")
                            .font(.system(size: 20, weight: .medium))
                            .foregroundColor(AppTheme.textSecondary)
                    }
                    
                    Spacer()
                    
                    VStack(spacing: 16) {
                        FintechButton(title: "Регистрация", isLoading: false, isDisabled: false) {
                            showingRegister = true
                        }
                        
                        // РАБОЧАЯ КНОПКА ВХОДА
                        NavigationLink(destination: LoginView(isLoggedIn: $isLoggedIn)) {
                            StubButton(title: "Вход")
                        }
                    }
                    .padding(.bottom, 50)
                }
                .padding(.horizontal, 24)
            }
            .sheet(isPresented: $showingRegister) {
                RegisterView(isLoggedIn: $isLoggedIn)
            }
        }
    }
}
