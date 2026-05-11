import SwiftUI

struct TransactionDetailView: View {
    let transaction: Transaction
    @Bindable var viewModel: TransactionViewModel
    @Environment(\.dismiss) var dismiss
    @State private var showingCategoryPicker = false
    
    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()
            
            VStack(spacing: 24) {
                VStack(spacing: 8) {
                    Text(transaction.isIncome ? "Доход" : "Расход")
                        .font(.caption).bold()
                        .foregroundColor(AppTheme.textSecondary)
                    
                    Text("\(transaction.isIncome ? "+" : "-")\(Double(transaction.amount) / 100.0, specifier: "%.2f") \(transaction.currency)")
                        .font(.system(size: 40, weight: .bold, design: .rounded))
                        .foregroundColor(transaction.isIncome ? AppTheme.accent : .white)
                }
                .padding(.vertical, 30)
                
                VStack(spacing: 1) {
                    DetailRow(title: "Название", value: transaction.name)
                    
                    Button {
                        showingCategoryPicker = true
                    } label: {
                        HStack {
                            Text("Категория").foregroundColor(AppTheme.textSecondary)
                            Spacer()
                            Text(viewModel.categories[transaction.categoryId ?? ""]?.nameCategory ?? "Не указана")
                                .foregroundColor(AppTheme.accent)
                            Image(systemName: "chevron.right").font(.caption)
                        }
                        .padding()
                        .background(AppTheme.bgSecondary)
                    }
                    
                    DetailRow(title: "Дата", value: transaction.completedAt.formatted(date: .long, time: .shortened))
                    DetailRow(title: "Тип", value: transaction.isImported == true ? "Импорт из PDF" : "Ручной ввод")
                    
                    if let extId = transaction.externalTransactionId {
                        DetailRow(title: "ID банка", value: String(extId.prefix(12)) + "...")
                    }
                }
                .cornerRadius(16)
                .padding(.horizontal)
                
                Spacer()
            }
        }
        .navigationTitle("Операция")
        .sheet(isPresented: $showingCategoryPicker) {
            CategoryPickerView(transaction: transaction, viewModel: viewModel)
        }
    }
}

struct DetailRow: View {
    let title: String
    let value: String
    var body: some View {
        HStack {
            Text(title).foregroundColor(AppTheme.textSecondary)
            Spacer()
            Text(value).foregroundColor(AppTheme.textPrimary)
        }
        .padding()
        .background(AppTheme.bgSecondary)
    }
}

struct CategoryPickerView: View {
    let transaction: Transaction
    let viewModel: TransactionViewModel
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground.ignoresSafeArea()
                List(Array(viewModel.categories.values).filter { $0.isIncome == transaction.isIncome }) { category in
                    Button {
                        Task {
                            let success = await viewModel.updateCategory(for: transaction, newCategoryId: category.categoryId)
                            if success { dismiss() }
                        }
                    } label: {
                        HStack {
                            Text(category.nameCategory)
                                .foregroundColor(.white)
                            Spacer()
                            if transaction.categoryId == category.categoryId {
                                Image(systemName: "checkmark").foregroundColor(AppTheme.accent)
                            }
                        }
                    }
                    .listRowBackground(AppTheme.bgSecondary)
                }
                .scrollContentBackground(.hidden)
            }
            .navigationTitle("Выберите категорию")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}
