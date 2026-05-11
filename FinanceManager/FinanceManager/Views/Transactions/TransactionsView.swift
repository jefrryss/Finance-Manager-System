import SwiftUI

struct TransactionsView: View {
    @State private var viewModel = TransactionViewModel()
    @State private var showingAddTransaction = false
    @State private var showingFilters = false
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.bgPrimary.ignoresSafeArea()
                
                VStack(spacing: 0) {
                    HStack {
                        Text("Операции")
                            .font(.largeTitle)
                            .bold()
                            .foregroundColor(.white)
                        Spacer()
                        Button {
                            showingAddTransaction = true
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.system(size: 28))
                                .foregroundColor(AppTheme.accent)
                        }
                    }
                    .padding(.horizontal, 20)
                    .padding(.top, 10)
                    
                    VStack(spacing: 16) {
                        HStack(spacing: 12) {
                            HStack {
                                Image(systemName: "magnifyingglass")
                                    .foregroundColor(.gray)
                                ZStack(alignment: .leading) {
                                    if viewModel.searchText.isEmpty {
                                        Text("Поиск...")
                                            .foregroundColor(Color.white.opacity(0.75))
                                    }
                                    TextField("", text: $viewModel.searchText)
                                        .foregroundColor(.white)
                                        .tint(AppTheme.accent)
                                }
                            }
                            .padding(12)
                            .background(AppTheme.bgSecondary)
                            .cornerRadius(12)
                            
                            Button {
                                showingFilters = true
                            } label: {
                                Image(systemName: "line.3.horizontal.decrease.circle")
                                    .font(.system(size: 20))
                                    .padding(12)
                                    .background((viewModel.selectedCategoryId != nil || !viewModel.dateInputText.isEmpty) ? AppTheme.accent : AppTheme.bgSecondary)
                                    .foregroundColor(.white)
                                    .cornerRadius(12)
                            }
                        }
                        .padding(.horizontal, 20)
                        
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 10) {
                                FilterChip(title: "Все", isSelected: viewModel.selectedFilterType == .all) {
                                    viewModel.selectedFilterType = .all
                                }
                                FilterChip(title: "Доходы", isSelected: viewModel.selectedFilterType == .income) {
                                    viewModel.selectedFilterType = .income
                                }
                                FilterChip(title: "Расходы", isSelected: viewModel.selectedFilterType == .expense) {
                                    viewModel.selectedFilterType = .expense
                                }
                            }
                            .padding(.horizontal, 20)
                        }
                        
                        if viewModel.isLoading && viewModel.transactions.isEmpty {
                            Spacer()
                            ProgressView()
                            Spacer()
                        } else if viewModel.filteredTransactions.isEmpty {
                            Spacer()
                            VStack(spacing: 12) {
                                Image(systemName: "tray")
                                    .font(.system(size: 40))
                                    .foregroundColor(AppTheme.textSecondary.opacity(0.5))
                                Text("Операций не найдено")
                                    .foregroundColor(AppTheme.textSecondary)
                            }
                            Spacer()
                        } else {
                            ScrollView {
                                LazyVStack(alignment: .leading, spacing: 16) {
                                    ForEach(viewModel.groupedTransactions, id: \.date) { group in
                                        VStack(alignment: .leading, spacing: 12) {
                                            Text(group.date)
                                                .font(.system(size: 14, weight: .bold))
                                                .foregroundColor(AppTheme.textSecondary)
                                                .padding(.horizontal, 20)
                                            
                                            VStack(spacing: 4) {
                                                ForEach(group.transactions) { transaction in
                                                    NavigationLink(destination: TransactionDetailView(transaction: transaction, viewModel: viewModel)) {
                                                        TransactionRow(transaction: transaction, category: viewModel.categories[transaction.categoryId ?? ""])
                                                            .padding(.horizontal, 20)
                                                    }
                                                    .buttonStyle(.plain)
                                                }
                                            }
                                        }
                                        .padding(.bottom, 8)
                                    }
                                }
                            }
                            .refreshable {
                                await viewModel.fetchTransactions()
                            }
                        }
                    }
                }
                .toolbar(.hidden, for: .navigationBar)
                .sheet(isPresented: $showingFilters) {
                    FilterSheetView(viewModel: viewModel)
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
    
    struct FilterChip: View {
        let title: String
        let isSelected: Bool
        let action: () -> Void
        
        var body: some View {
            Button(action: action) {
                Text(title)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(isSelected ? .white : AppTheme.textSecondary)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 10)
                    .background(isSelected ? AppTheme.accent : AppTheme.bgSecondary.opacity(0.7))
                    .cornerRadius(12)
                    .overlay(
                        RoundedRectangle(cornerRadius: 12)
                            .stroke(isSelected ? AppTheme.accent : Color.white.opacity(0.15), lineWidth: 1)
                    )
            }
        }
    }
}
