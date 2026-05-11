import Foundation
import Observation

@Observable
class TransactionViewModel {
    var transactions: [Transaction] = []
    var categories: [String: TransactionCategory] = [:]
    var isLoading = false
    var errorMessage: String?
    
    var searchText = ""
    var selectedFilterType: TransactionFilterType = .all
    var selectedAccountId: String?
    var selectedCategoryId: String?
    
    var dateInputText = ""
    var categorySearchText = ""
    
    enum TransactionFilterType {
        case all
        case income
        case expense
    }
    
    var groupedTransactions: [(date: String, transactions: [Transaction])] {
        let filtered = filteredTransactions
        var grouped: [String: [Transaction]] = [:]
        
        for transaction in filtered {
            let dateKey = dateGroupKey(transaction.completedAt)
            if grouped[dateKey] == nil {
                grouped[dateKey] = []
            }
            grouped[dateKey]?.append(transaction)
        }
        
        return grouped.map { (date: $0.key, transactions: $0.value) }
            .sorted { group1, group2 in
                let date1 = group1.transactions.first?.completedAt ?? Date.distantPast
                let date2 = group2.transactions.first?.completedAt ?? Date.distantPast
                return date1 > date2
            }
    }
    
    var filteredTransactions: [Transaction] {
        var result = transactions
        let calendar = Calendar.current
        
        switch selectedFilterType {
        case .income:
            result = result.filter { $0.isIncome }
        case .expense:
            result = result.filter { !$0.isIncome }
        case .all:
            break
        }
        
        if !dateInputText.isEmpty {
            result = result.filter { tx in
                let monthName = tx.completedAt.formatted(.dateTime.month(.wide)).lowercased()
                let yearString = String(calendar.component(.year, from: tx.completedAt))
                let input = dateInputText.lowercased()
                return monthName.contains(input) || yearString.contains(input)
            }
        }
        
        if let categoryId = selectedCategoryId {
            result = result.filter { $0.categoryId == categoryId }
        }
        
        if !searchText.isEmpty {
            result = result.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        
        return result
    }
    
    var searchedCategories: [TransactionCategory] {
        let allCats = Array(categories.values)
        if categorySearchText.isEmpty {
            return allCats.sorted { $0.nameCategory < $1.nameCategory }
        }
        return allCats.filter { $0.nameCategory.localizedCaseInsensitiveContains(categorySearchText) }
            .sorted { $0.nameCategory < $1.nameCategory }
    }
    
    var statistics: (income: Double, expense: Double) {
        let filtered = filteredTransactions
        let income = Double(filtered.filter { $0.isIncome }.reduce(0) { $0 + $1.amount }) / 100.0
        let expense = Double(filtered.filter { !$0.isIncome }.reduce(0) { $0 + $1.amount }) / 100.0
        return (income, expense)
    }
    
    func fetchTransactions() async {
        isLoading = true
        errorMessage = nil
        do {
            async let fetchedAccounts: [Account] = try NetworkManager.shared.fetch(endpoint: "/accounts")
            async let fetchedCategories: [TransactionCategory] = try NetworkManager.shared.fetch(endpoint: "/categories")
            async let fetchedTransactions: [Transaction] = try NetworkManager.shared.fetch(endpoint: "/transactions")
            
            let (accs, cats, txs) = try await (fetchedAccounts, fetchedCategories, fetchedTransactions)
            let activeAccountIds = Set(accs.map { $0.accountId })
            
            self.categories = Dictionary(uniqueKeysWithValues: cats.map { ($0.categoryId, $0) })
            self.transactions = txs.filter { activeAccountIds.contains($0.accountId) }
        } catch {
            self.errorMessage = error.localizedDescription
        }
        isLoading = false
    }
    
    func updateCategory(for transaction: Transaction, newCategoryId: String) async -> Bool {
        let body = ["category_id": newCategoryId]
        do {
            let _: [String: String] = try await NetworkManager.shared.post(endpoint: "/transactions/\(transaction.transactionId)", body: body)
            await fetchTransactions()
            return true
        } catch {
            return false
        }
    }
    
    private func dateGroupKey(_ date: Date) -> String {
        let calendar = Calendar.current
        let now = Date()
        
        if calendar.isDateInToday(date) {
            return "Сегодня"
        } else if calendar.isDateInYesterday(date) {
            return "Вчера"
        } else if calendar.isDate(date, equalTo: now, toGranularity: .weekOfYear) {
            return "На этой неделе"
        } else if calendar.isDate(date, equalTo: now, toGranularity: .year) {
            return date.formatted(.dateTime.month(.wide).year())
        } else {
            return date.formatted(.dateTime.month(.wide).year())
        }
    }
}
