import Foundation
import Observation

@Observable
class TransactionViewModel {
    var transactions: [Transaction] = []
    var isLoading = false
    var errorMessage: String?
    
    var searchText = ""
    var selectedFilterType: TransactionFilterType = .all
    var selectedAccountId: String?
    
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
        
        return grouped.sorted { $0.key > $1.key }.map { (date: $0.key, transactions: $0.value) }
    }
    
    var filteredTransactions: [Transaction] {
        var result = transactions
        
        switch selectedFilterType {
        case .income:
            result = result.filter { $0.isIncome }
        case .expense:
            result = result.filter { !$0.isIncome }
        case .all:
            break
        }
        
        if let accountId = selectedAccountId {
            result = result.filter { $0.accountId == accountId }
        }
        
        if !searchText.isEmpty {
            result = result.filter { $0.name.localizedCaseInsensitiveContains(searchText) }
        }
        
        return result.sorted { $0.completedAt > $1.completedAt }
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
            let activeAccounts: [Account] = try await NetworkManager.shared.fetch(endpoint: "/accounts")
            let activeAccountIds = Set(activeAccounts.map { $0.accountId })
            
            let allTransactions: [Transaction] = try await NetworkManager.shared.fetch(endpoint: "/transactions")
            
            self.transactions = allTransactions.filter { activeAccountIds.contains($0.accountId) }
            
            print("✅ Транзакции загружены: \(self.transactions.count)")
        } catch {
            self.errorMessage = error.localizedDescription
            print("❌ Ошибка загрузки транзакций: \(error)")
        }
        isLoading = false
    }
    
    func fetchTransactionsForAccount(_ accountId: String) async {
        isLoading = true
        errorMessage = nil
        do {
            let allTransactions: [Transaction] = try await NetworkManager.shared.fetch(endpoint: "/transactions")
            self.transactions = allTransactions.filter { $0.accountId == accountId }
            print("✅ Транзакции счета \(accountId) загружены: \(self.transactions.count)")
        } catch {
            self.errorMessage = error.localizedDescription
            print("❌ Ошибка загрузки транзакций счета: \(error)")
        }
        isLoading = false
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
