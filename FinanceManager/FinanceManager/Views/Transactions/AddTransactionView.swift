import SwiftUI

struct AddTransactionView: View {
    @Environment(\.dismiss) var dismiss
    @State private var viewModel = AddTransactionViewModel()
    var onSave: () -> Void
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground.ignoresSafeArea()
                
                ScrollView {
                    VStack(spacing: 24) {
                        Picker("Тип", selection: $viewModel.isIncome) {
                            Text("Расход").tag(false)
                            Text("Доход").tag(true)
                        }
                        .pickerStyle(.segmented)
                        .colorMultiply(AppTheme.accent)
                        .onChange(of: viewModel.isIncome) { _, _ in
                            viewModel.updateCategorySelection()
                        }
                        
                        VStack(spacing: 16) {
                            FintechTextField(icon: "pencil", placeholder: "Название (напр. Кофе)", text: $viewModel.name)
                            FintechTextField(icon: "rublesign", placeholder: "Сумма", text: $viewModel.amountString)
                                .keyboardType(.decimalPad)
                        }
                        
                        VStack(alignment: .leading, spacing: 8) {
                            Text("Счет").foregroundColor(AppTheme.textSecondary).font(.caption)
                            Picker("Счет", selection: $viewModel.selectedAccountId) {
                                ForEach(viewModel.accounts) { acc in
                                    Text(acc.nameAccount).tag(acc.accountId as String?)
                                }
                            }
                            .tint(AppTheme.textPrimary)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding()
                            .background(AppTheme.bgSecondary)
                            .cornerRadius(20)
                        }
                        
                        VStack(alignment: .leading, spacing: 16) {
                            HStack {
                                Text("Категория").foregroundColor(AppTheme.textSecondary).font(.caption)
                                Spacer()
                                Toggle("Своя", isOn: $viewModel.isCustomCategory)
                                    .labelsHidden()
                                    .tint(AppTheme.accent)
                            }
                            
                            if viewModel.isCustomCategory {
                                VStack(alignment: .leading, spacing: 4) {
                                    FintechTextField(icon: "tag", placeholder: "Название категории", text: $viewModel.customCategoryName)
                                    if viewModel.categories.isEmpty {
                                        Text("База категорий пуста, введите название вручную")
                                            .font(.system(size: 10))
                                            .foregroundColor(.orange)
                                            .padding(.leading, 4)
                                    }
                                }
                            } else {
                                Picker("Выберите категорию", selection: $viewModel.selectedCategoryId) {
                                    ForEach(viewModel.categories.filter { $0.isIncome == viewModel.isIncome }) { cat in
                                        Text(cat.nameCategory).tag(cat.categoryId as String?)
                                    }
                                }
                                .tint(AppTheme.textPrimary)
                                .frame(maxWidth: .infinity, alignment: .leading)
                                .padding()
                                .background(AppTheme.bgSecondary)
                                .cornerRadius(20)
                            }
                        }
                        
                        if let error = viewModel.errorMessage {
                            Text(error).foregroundColor(.red).font(.caption)
                        }
                        
                        Spacer(minLength: 20)
                        
                        FintechButton(title: "Добавить", isLoading: viewModel.isLoading, isDisabled: viewModel.name.isEmpty || viewModel.amountString.isEmpty) {
                            Task {
                                if await viewModel.saveTransaction() {
                                    onSave()
                                    dismiss()
                                }
                            }
                        }
                    }
                    .padding(24)
                }
            }
            .navigationTitle(viewModel.isIncome ? "Новый доход" : "Новый расход")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Отмена") { dismiss() }.foregroundColor(AppTheme.textSecondary)
                }
            }
            .task {
                await viewModel.fetchFormOptions()
            }
        }
    }
}
