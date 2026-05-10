import SwiftUI

struct ProfileView: View {
    @AppStorage("isLoggedIn") private var isLoggedIn = true
    @State private var viewModel = ProfileViewModel()
    @State private var showSignOutAlert = false
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground.ignoresSafeArea()
                
                ScrollView {
                    VStack(spacing: 24) {
                        VStack(spacing: 12) {
                            ZStack {
                                Circle()
                                    .fill(AppTheme.accent.opacity(0.1))
                                    .frame(width: 100, height: 100)
                                
                                Image(systemName: "person.fill")
                                    .font(.system(size: 40))
                                    .foregroundColor(AppTheme.accent)
                            }
                            
                            Text("Ваш профиль")
                                .font(.system(size: 24, weight: .bold))
                                .foregroundColor(AppTheme.textPrimary)
                        }
                        .padding(.top, 20)
                        
                        VStack(spacing: 1) {
                            ProfileRow(icon: "lock.fill", title: "Сменить пароль", color: .blue) {
                            }
                            
                            ProfileRow(icon: "at", title: "Сменить логин", color: .purple) {
                            }
                        }
                        .background(AppTheme.bgSecondary.opacity(0.5))
                        .cornerRadius(20)
                        .padding(.horizontal, 24)
                        
                        VStack(alignment: .leading, spacing: 16) {
                            Text("Данные")
                                .font(.system(size: 14, weight: .semibold))
                                .foregroundColor(AppTheme.textSecondary)
                                .padding(.horizontal, 40)
                            
                            VStack(spacing: 1) {
                                NavigationLink(destination: Text("Экран категорий")) {
                                    ProfileRow(icon: "tag.fill", title: "Мои категории", color: AppTheme.accent)
                                }
                            }
                            .background(AppTheme.bgSecondary.opacity(0.5))
                            .cornerRadius(20)
                            .padding(.horizontal, 24)
                        }
                        
                        Button {
                            showSignOutAlert = true
                        } label: {
                            HStack {
                                Image(systemName: "rectangle.portrait.and.arrow.right")
                                Text("Выйти из аккаунта")
                                    .fontWeight(.semibold)
                            }
                            .foregroundColor(.red)
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.red.opacity(0.1))
                            .cornerRadius(16)
                        }
                        .padding(.horizontal, 24)
                        .padding(.top, 20)
                    }
                    .padding(.bottom, 30)
                }
            }
            .navigationTitle("Настройки")
            .navigationBarTitleDisplayMode(.inline)
            .alert("Выход", isPresented: $showSignOutAlert) {
                Button("Отмена", role: .cancel) { }
                Button("Выйти", role: .destructive) {
                    viewModel.logout()
                    isLoggedIn = false
                }
            } message: {
                Text("Вы уверены, что хотите выйти? Вам придется снова вводить логин и пароль.")
            }
        }
    }
}

// Вспомогательный компонент для строк меню
struct ProfileRow: View {
    let icon: String
    let title: String
    let color: Color
    var action: (() -> Void)? = nil
    
    var body: some View {
        Button {
            action?()
        } label: {
            HStack(spacing: 16) {
                ZStack {
                    RoundedRectangle(cornerRadius: 8)
                        .fill(color.opacity(0.2))
                        .frame(width: 32, height: 32)
                    Image(systemName: icon)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(color)
                }
                
                Text(title)
                    .foregroundColor(AppTheme.textPrimary)
                
                Spacer()
                
                Image(systemName: "chevron.right")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(AppTheme.textSecondary.opacity(0.5))
            }
            .padding()
        }
    }
}
