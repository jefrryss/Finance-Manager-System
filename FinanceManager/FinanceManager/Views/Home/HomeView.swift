import SwiftUI

struct HomeView: View {
    @State private var viewModel = AccountViewModel()
    @State private var txViewModel = TransactionViewModel()
    @State private var showingAddTransaction = false
    @State private var isCreatingIncome = false
    @State private var showingAddAccount = false
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground.ignoresSafeArea()
                
                ScrollView(showsIndicators: false) {
                    VStack(alignment: .leading, spacing: 32) {
                        
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Ваш капитал")
                                .font(.system(size: 16, weight: .medium))
                                .foregroundColor(AppTheme.textSecondary)
                            
                            Text("\(viewModel.totalBalance, specifier: "%.2f") ₽")
                                .font(.system(size: 44, weight: .bold, design: .rounded))
                                .foregroundColor(AppTheme.textPrimary)
                        }
                        .padding(.horizontal, 24)
                        .padding(.top, 20)
                        
                        HStack(spacing: 20) {
                            QuickActionButton(title: "Трата", icon: "minus.circle.fill", color: .red) {
                                isCreatingIncome = false
                                showingAddTransaction = true
                            }
                            QuickActionButton(title: "Доход", icon: "plus.circle.fill", color: AppTheme.accent) {
                                isCreatingIncome = true
                                showingAddTransaction = true
                            }
                            QuickActionButton(title: "Лимиты", icon: "gauge.with.needle.fill", color: .blue) {
                            }
                        }
                        .padding(.horizontal, 24)
                        
                        VStack(alignment: .leading, spacing: 16) {
                            HStack {
                                Text("Мои счета")
                                    .font(.system(size: 22, weight: .bold))
                                    .foregroundColor(AppTheme.textPrimary)
                                Spacer()
                                
                                Button {
                                    showingAddAccount = true
                                } label: {
                                    Image(systemName: "plus.circle.fill")
                                        .foregroundColor(AppTheme.accent)
                                        .font(.title2)
                                }
                            }
                            .padding(.horizontal, 24)
                            
                            ScrollView(.horizontal, showsIndicators: false) {
                                HStack(spacing: 16) {
                                    if viewModel.accounts.isEmpty {
                                        Text("У вас пока нет счетов")
                                            .foregroundColor(AppTheme.textSecondary)
                                            .padding(.leading, 24)
                                    } else {
                                        ForEach(viewModel.accounts, id: \.accountId) { account in
                                            NavigationLink(destination: AccountDetailView(account: account)) {
                                                AccountListRow(account: account)
                                                    .frame(width: 280)
                                            }
                                            .buttonStyle(PlainButtonStyle())
                                        }
                                        .padding(.horizontal, 24)
                                    }
                                }
                            }
                        }
                        
                        VStack(alignment: .leading, spacing: 16) {
                            HStack {
                                Text("История")
                                    .font(.system(size: 22, weight: .bold))
                                    .foregroundColor(AppTheme.textPrimary)
                                Spacer()
                                NavigationLink("Все", destination: TransactionsView())
                                    .font(.system(size: 14))
                                    .foregroundColor(AppTheme.accent)
                            }
                            .padding(.horizontal, 24)
                            
                            if txViewModel.isLoading && txViewModel.transactions.isEmpty {
                            ProgressView()
                                .frame(maxWidth: .infinity, alignment: .center)
                                .padding(.vertical, 20)
                        } else {
                            TransactionsPreviewList(transactions: txViewModel.transactions)
                        }
                        }
                        
                        Spacer(minLength: 100)
                    }
                }
            }
            .navigationBarTitleDisplayMode(.inline)
            .sheet(isPresented: $showingAddAccount) {
                AddAccountView(onSave: {
                    Task {
                        await viewModel.fetchAccounts()
                    }
                })
            }
            .sheet(isPresented: $showingAddTransaction) {
                AddTransactionView(initialIsIncome: isCreatingIncome) {
                    Task {
                        await viewModel.fetchAccounts()
                        await txViewModel.fetchTransactions()
                    }
                }
            }
            .task {
                await viewModel.fetchAccounts()
                await txViewModel.fetchTransactions()
            }
        }
    }
}

struct QuickActionButton: View {
    let title: String
    let icon: String
    let color: Color
    var action: () -> Void
    
    var body: some View {
        Button(action: action) {
            VStack(spacing: 10) {
                Image(systemName: icon)
                    .font(.system(size: 24))
                    .foregroundColor(color)
                Text(title)
                    .font(.system(size: 14, weight: .medium))
                    .foregroundColor(AppTheme.textPrimary)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(AppTheme.bgSecondary.opacity(0.8))
            .cornerRadius(20)
            .overlay(
                RoundedRectangle(cornerRadius: 20)
                    .stroke(Color.white.opacity(0.05), lineWidth: 1)
            )
        }
    }
}
