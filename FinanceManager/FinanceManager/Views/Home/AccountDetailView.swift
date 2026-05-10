import SwiftUI

struct AccountDetailView: View {
    let account: Account
    @Environment(\.dismiss) var dismiss
    @State private var viewModel = AccountViewModel()
    @State private var showingDeleteAlert = false
    
    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()
            
            VStack(spacing: 24) {
                VStack(spacing: 16) {
                    Circle()
                        .fill(Color(hex: account.colorHex))
                        .frame(width: 60, height: 60)
                        .overlay(Image(systemName: "creditcard.fill").foregroundColor(.white))
                    
                    Text(account.nameAccount)
                        .font(.title2).bold()
                        .foregroundColor(AppTheme.textPrimary)
                    
                    Text("\(Double(account.balance) / 100.0, specifier: "%.2f") \(account.currency)")
                        .font(.system(size: 34, weight: .bold, design: .rounded))
                        .foregroundColor(AppTheme.textPrimary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 40)
                .background(AppTheme.bgSecondary)
                .cornerRadius(24)
                
                Spacer()
                
                Button(role: .destructive) {
                    showingDeleteAlert = true
                } label: {
                    HStack {
                        Image(systemName: "trash")
                        Text("Удалить счет")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.red.opacity(0.1))
                    .cornerRadius(16)
                }
            }
            .padding(24)
        }
        .navigationTitle("Детали счета")
        .alert("Удалить счет?", isPresented: $showingDeleteAlert) {
            Button("Отмена", role: .cancel) { }
            Button("Удалить", role: .destructive) {
                Task {
                    let success = await viewModel.deleteAccount(account)
                    if success { dismiss() }
                }
            }
        } message: {
            Text("Это действие нельзя будет отменить. Все данные счета будут заархивированы.")
        }
    }
}
