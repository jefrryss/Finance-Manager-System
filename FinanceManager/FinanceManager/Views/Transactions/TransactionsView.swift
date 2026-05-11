import SwiftUI

struct TransactionsView: View {
    @State private var viewModel = TransactionViewModel()
    @State private var showingAddTransaction = false
    @State private var accountViewModel = AccountViewModel()
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.bgPrimary.ignoresSafeArea()
                
                VStack(spacing: 0) {
                    // Статистика
                    if !viewModel.filteredTransactions.isEmpty {
                        HStack(spacing: 20) {
                            VStack(alignment: .leading, spacing: 4) {
                                Text("Доходы")
                                    .font(.caption)
                                    .foregroundColor(AppTheme.textSecondary)
                                Text("+\(viewModel.statistics.income, specifier: "%.2f") ₽")
                                    .font(.system(size: 16, weight: .semibold))
                                    .foregroundColor(AppTheme.accent)
                            }
                            
                            Divider()
                                .frame(height: 40)
                            
                            VStack(alignment: .leading, spacing: 4) {
                                Text("Расходы")
                                    .font(.caption)
                                    .foregroundColor(AppTheme.textSecondary)
                                Text("-\(viewModel.statistics.expense, specifier: "%.2f") ₽")
                                    .font(.system(size: 16, weight: .semibold))
                                    .foregroundColor(.red)
                            }
                            
                            Spacer()
                        }
                        .padding(16)
                        .background(AppTheme.bgSecondary)
                        .cornerRadius(12)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                    }
                    
                    // Поиск и фильтры
                    VStack(spacing: 12) {
                        HStack {
                            Image(systemName: "magnifyingglass")
                                .foregroundColor(AppTheme.textSecondary)
                            
                            TextField("Поиск операции", text: $viewModel.searchText)
                                .textFieldStyle(.plain)
                                .foregroundColor(AppTheme.textPrimary)
                        }
                        .padding(12)
                        .background(AppTheme.bgSecondary)
                        .cornerRadius(10)
                        
                        HStack(spacing: 10) {
                            FilterChip(
                                title: "Все",
                                isSelected: viewModel.selectedFilterType == .all,
                                action: { viewModel.selectedFilterType = .all }
                            )
                            
                            FilterChip(
                                title: "Доходы",
                                isSelected: viewModel.selectedFilterType == .income,
                                action: { viewModel.selectedFilterType = .income }
                            )
                            
                            FilterChip(
                                title: "Расходы",
                                isSelected: viewModel.selectedFilterType == .expense,
                                action: { viewModel.selectedFilterType = .expense }
                            )
                            
                            Spacer()
                        }
                    }
                    .padding(16)
                    
                    // Операции
                    if viewModel.isLoading && viewModel.transactions.isEmpty {
                        Spacer()
                        ProgressView()
                        Spacer()
                    } else if let error = viewModel.errorMessage {
                        Spacer()
                        VStack(spacing: 12) {
                            Image(systemName: "exclamationmark.circle")
                                .font(.system(size: 40))
                                .foregroundColor(.red)
                            Text(error)
                                .foregroundColor(.red)
                                .multilineTextAlignment(.center)
                        }
                        Spacer()
                    } else if viewModel.filteredTransactions.isEmpty {
                        Spacer()
                        VStack(spacing: 12) {
                            Image(systemName: "list.bullet")
                                .font(.system(size: 40))
                                .foregroundColor(AppTheme.textSecondary)
                            Text("Операций не найдено")
                                .foregroundColor(AppTheme.textSecondary)
                        }
                        Spacer()
                    } else {
                        ScrollView {
                            LazyVStack(alignment: .leading, spacing: 12) {
                                ForEach(viewModel.groupedTransactions, id: \.date) { group in
                                    VStack(alignment: .leading, spacing: 8) {
                                        Text(group.date)
                                            .font(.system(size: 14, weight: .semibold))
                                            .foregroundColor(AppTheme.textSecondary)
                                            .padding(.horizontal, 16)
                                        
                                        ForEach(group.transactions) { transaction in
                                            TransactionRow(transaction: transaction)
                                        }
                                    }
                                }
                            }
                            .padding(.vertical, 8)
                        }
                        .refreshable {
                            await viewModel.fetchTransactions()
                        }
                    }
                }
            }
            .navigationTitle("Операции")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                        showingAddTransaction = true
                    } label: {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 22))
                            .foregroundColor(AppTheme.accent)
                    }
                }
            }
            .sheet(isPresented: $showingAddTransaction) {
                AddTransactionView {
                    Task {
                        await viewModel.fetchTransactions()
                    }
                }
            }
            .task {
                await viewModel.fetchTransactions()
            }
        }
    }
}

// Компонент фильтр-чипа
struct FilterChip: View {
    let title: String
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(title)
                .font(.system(size: 13, weight: .semibold))
                .foregroundColor(isSelected ? .white : AppTheme.textSecondary)
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(isSelected ? AppTheme.accent : AppTheme.bgSecondary)
                .cornerRadius(8)
        }
    }
}
