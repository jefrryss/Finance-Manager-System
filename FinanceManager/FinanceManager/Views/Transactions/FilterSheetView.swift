import SwiftUI

struct FilterSheetView: View {
    @Bindable var viewModel: TransactionViewModel
    @Environment(\.dismiss) var dismiss
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.bgPrimary.ignoresSafeArea()
                
                VStack(spacing: 24) {
                    VStack(alignment: .leading, spacing: 14) {
                        Text("Период")
                            .font(.system(size: 18, weight: .bold))
                            .foregroundColor(.white)
                        
                        ZStack(alignment: .leading) {
                            if viewModel.dateInputText.isEmpty {
                                Text("Введите месяц или год (напр. Май)")
                                    .foregroundColor(Color.white.opacity(0.75))
                            }
                            TextField("", text: $viewModel.dateInputText)
                                .foregroundColor(.white)
                                .tint(AppTheme.accent)
                        }
                        .padding(16)
                        .background(AppTheme.bgSecondary)
                        .cornerRadius(12)
                        .overlay(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(Color.white.opacity(0.1), lineWidth: 1)
                        )
                        
                        HStack(spacing: 10) {
                            QuickDateButton(title: "2026", viewModel: viewModel)
                            QuickDateButton(title: "Май", viewModel: viewModel)
                            Spacer()
                            QuickDateButton(title: "Сбросить", viewModel: viewModel, isReset: true)
                        }
                    }
                    .padding(.horizontal, 20)
                    
                    VStack(alignment: .leading, spacing: 14) {
                        Text("Категория")
                            .font(.system(size: 18, weight: .bold))
                            .foregroundColor(.white)
                        
                        HStack {
                            Image(systemName: "magnifyingglass")
                                .foregroundColor(.gray)
                            ZStack(alignment: .leading) {
                                if viewModel.categorySearchText.isEmpty {
                                    Text("Поиск категории...")
                                        .foregroundColor(Color.white.opacity(0.75))
                                }
                                TextField("", text: $viewModel.categorySearchText)
                                    .foregroundColor(.white)
                                    .tint(AppTheme.accent)
                            }
                        }
                        .padding(14)
                        .background(AppTheme.bgSecondary)
                        .cornerRadius(12)
                        .overlay(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(Color.white.opacity(0.1), lineWidth: 1)
                        )
                        
                        ScrollView {
                            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                                ForEach(viewModel.searchedCategories) { cat in
                                    CategoryChip(category: cat, isSelected: viewModel.selectedCategoryId == cat.categoryId) {
                                        viewModel.selectedCategoryId = (viewModel.selectedCategoryId == cat.categoryId) ? nil : cat.categoryId
                                    }
                                }
                            }
                            .padding(.vertical, 4)
                        }
                        .frame(maxHeight: 280)
                    }
                    .padding(.horizontal, 20)
                    
                    Spacer()
                    
                    Button(action: { dismiss() }) {
                        Text("Показать результаты")
                            .font(.system(size: 16, weight: .bold))
                            .foregroundColor(.white)
                            .frame(maxWidth: .infinity)
                            .padding(.vertical, 16)
                            .background(AppTheme.accent)
                            .cornerRadius(16)
                            .shadow(color: AppTheme.accent.opacity(0.3), radius: 10, x: 0, y: 5)
                    }
                    .padding(.horizontal, 20)
                    .padding(.bottom, 20)
                }
                .padding(.top, 20)
            }
            .navigationTitle("Фильтры")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Сбросить всё") {
                        viewModel.selectedCategoryId = nil
                        viewModel.dateInputText = ""
                        viewModel.selectedFilterType = .all
                    }
                    .foregroundColor(AppTheme.accent)
                }
            }
        }
    }
}

struct QuickDateButton: View {
    let title: String
    var viewModel: TransactionViewModel
    var isReset = false
    
    var body: some View {
        Button(action: {
            viewModel.dateInputText = isReset ? "" : title
        }) {
            Text(title)
                .font(.system(size: 13, weight: .bold))
                .padding(.horizontal, 16)
                .padding(.vertical, 10)
                .background(isReset ? Color.red.opacity(0.1) : AppTheme.bgSecondary)
                .foregroundColor(isReset ? .red : AppTheme.accent)
                .cornerRadius(10)
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(isReset ? Color.red.opacity(0.2) : Color.white.opacity(0.05), lineWidth: 1)
                )
        }
    }
}

struct CategoryChip: View {
    let category: TransactionCategory
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(category.nameCategory)
                .font(.system(size: 13, weight: .medium))
                .foregroundColor(isSelected ? .white : AppTheme.textSecondary)
                .frame(maxWidth: .infinity)
                .padding(.vertical, 12)
                .background(isSelected ? AppTheme.accent : AppTheme.bgSecondary.opacity(0.6))
                .cornerRadius(14)
                .overlay(
                    RoundedRectangle(cornerRadius: 14)
                        .stroke(isSelected ? AppTheme.accent : Color.white.opacity(0.1), lineWidth: 1)
                )
        }
    }
}
