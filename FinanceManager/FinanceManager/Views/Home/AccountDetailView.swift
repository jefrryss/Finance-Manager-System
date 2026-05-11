import SwiftUI

struct AccountDetailView: View {
    let account: Account
    @Environment(\.dismiss) var dismiss
    @State private var viewModel = AccountViewModel()
    @State private var txViewModel = TransactionViewModel()
    @State private var showingDeleteAlert = false
    
    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()
            
            ScrollView {
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
                    
                    if txViewModel.isLoading && txViewModel.filteredTransactions.isEmpty {
                        VStack(alignment: .leading, spacing: 12) {
                            Text("История операций")
                                .font(.system(size: 18, weight: .bold))
                                .foregroundColor(AppTheme.textPrimary)
                                .padding(.horizontal, 4)
                            
                            ProgressView()
                                .frame(maxWidth: .infinity, alignment: .center)
                                .padding(.vertical, 20)
                        }
                    } else if !txViewModel.filteredTransactions.isEmpty {
                        VStack(alignment: .leading, spacing: 12) {
                            Text("История операций")
                                .font(.system(size: 18, weight: .bold))
                                .foregroundColor(AppTheme.textPrimary)
                                .padding(.horizontal, 4)
                            
                            let recentTxs = Array(txViewModel.filteredTransactions.prefix(5))
                            
                            ForEach(recentTxs) { transaction in
                                TransactionRow(transaction: transaction)
                                    .padding(.vertical, 4)
                            }
                            
                            if txViewModel.filteredTransactions.count > 5 {
                                Text("Показаны последние 5 операций")
                                    .font(.caption)
                                    .foregroundColor(AppTheme.textSecondary)
                                    .frame(maxWidth: .infinity, alignment: .center)
                                    .padding(.top, 8)
                            }
                        }
                    } else {
                        VStack(spacing: 12) {
                            Image(systemName: "list.bullet")
                                .font(.system(size: 32))
                                .foregroundColor(AppTheme.textSecondary)
                            Text("Операций нет")
                                .foregroundColor(AppTheme.textSecondary)
                                .font(.system(size: 14, weight: .medium))
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 40)
                    }
                    
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
                    
                    Spacer(minLength: 40)
                }
                .padding(24)
            }
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
        .task {
            txViewModel.selectedAccountId = account.accountId
            await txViewModel.fetchTransactionsForAccount(account.accountId)
        }
    }
}
