import SwiftUI

struct CategoriesView: View {
    @State private var categories: [TransactionCategory] = []
    @State private var searchText = ""
    @State private var isLoading = false
    @State private var errorMessage: String?

    private var filteredCategories: [TransactionCategory] {
        let sorted = categories.sorted { $0.nameCategory < $1.nameCategory }
        guard !searchText.isEmpty else { return sorted }
        return sorted.filter { $0.nameCategory.localizedCaseInsensitiveContains(searchText) }
    }

    var body: some View {
        ZStack {
            AppTheme.finexaBackground.ignoresSafeArea()

            VStack(spacing: 16) {
                HStack {
                    Image(systemName: "magnifyingglass")
                        .foregroundColor(.gray)
                    ZStack(alignment: .leading) {
                        if searchText.isEmpty {
                            Text("Поиск категории")
                                .foregroundColor(Color.white.opacity(0.75))
                        }
                        TextField("", text: $searchText)
                            .foregroundColor(.white)
                            .tint(AppTheme.accent)
                    }
                }
                .padding(12)
                .background(AppTheme.bgSecondary)
                .cornerRadius(12)
                .padding(.horizontal, 24)

                if isLoading {
                    Spacer()
                    ProgressView()
                    Spacer()
                } else if let errorMessage {
                    Spacer()
                    Text(errorMessage)
                        .foregroundColor(.red)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal, 24)
                    Spacer()
                } else if filteredCategories.isEmpty {
                    Spacer()
                    VStack(spacing: 10) {
                        Image(systemName: "tray")
                            .font(.system(size: 40))
                            .foregroundColor(AppTheme.textSecondary.opacity(0.5))
                        Text("Категории не найдены")
                            .foregroundColor(AppTheme.textSecondary)
                    }
                    Spacer()
                } else {
                    ScrollView {
                        LazyVStack(spacing: 1) {
                            ForEach(filteredCategories) { category in
                                HStack {
                                    VStack(alignment: .leading, spacing: 4) {
                                        Text(category.nameCategory)
                                            .foregroundColor(AppTheme.textPrimary)
                                            .font(.system(size: 16, weight: .medium))
                                        Text(category.isIncome ? "Доход" : "Расход")
                                            .foregroundColor(AppTheme.textSecondary)
                                            .font(.system(size: 12))
                                    }
                                    Spacer()
                                }
                                .padding()
                                .background(AppTheme.bgSecondary.opacity(0.5))
                            }
                        }
                        .cornerRadius(16)
                        .padding(.horizontal, 24)
                    }
                }
            }
            .padding(.top, 20)
        }
        .navigationTitle("Мои категории")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await loadCategories()
        }
    }

    private func loadCategories() async {
        isLoading = true
        errorMessage = nil
        do {
            categories = try await NetworkManager.shared.fetch(endpoint: "/categories")
        } catch {
            errorMessage = "Не удалось загрузить категории"
        }
        isLoading = false
    }
}
