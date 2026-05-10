import SwiftUI

struct TransactionsView: View {
    @State private var viewModel = TransactionViewModel()
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.bgPrimary.ignoresSafeArea()
                
                if viewModel.isLoading && viewModel.transactions.isEmpty {
                    ProgressView()
                } else if let error = viewModel.errorMessage {
                    Text(error).foregroundColor(.red)
                } else {
                    ScrollView {
                        LazyVStack(spacing: 16) {
                            ForEach(viewModel.transactions) { transaction in
                                TransactionRow(transaction: transaction)
                            }
                        }
                        .padding(.horizontal, 20)
                        .padding(.top, 10)
                    }
                    .refreshable {
                        await viewModel.fetchTransactions()
                    }
                }
            }
            .navigationTitle("Операции")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button {
                    } label: {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 22))
                            .foregroundColor(AppTheme.accent)
                    }
                }
            }
            .task {
                await viewModel.fetchTransactions()
            }
        }
    }
}
