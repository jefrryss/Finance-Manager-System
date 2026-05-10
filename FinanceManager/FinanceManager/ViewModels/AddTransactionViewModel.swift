import Foundation
import Observation

@Observable
class AddTransactionViewModel {
    var accounts: [Account] = []
    var categories: [TransactionCategory] = []
    
    var selectedAccountId: UUID?
    var selectedCategoryId: UUID?
    var name = ""
    var amountString = ""
    var isIncome = false
    
    var isLoading = false
    var errorMessage: String?
    
    func fetchFormOptions() async {
        do {
            async let fetchedAccounts: [Account] = try NetworkManager.shared.fetch(endpoint: "/accounts")
            async let fetchedCategories: [TransactionCategory] = try NetworkManager.shared.fetch(endpoint: "/categories")
            
            let (accs, cats) = try await (fetchedAccounts, fetchedCategories)
            self.accounts = accs
            self.categories = cats
            
            if let firstAcc = accs.first { self.selectedAccountId = firstAcc.accountId }
            if let firstCat = cats.first { self.selectedCategoryId = firstCat.categoryId }
        } catch {
            self.errorMessage = "Не удалось загрузить данные"
        }
    }
    
    func saveTransaction() async -> Bool {
        guard let accId = selectedAccountId,
              let catId = selectedCategoryId,
              let amountDouble = Double(amountString.replacingOccurrences(of: ",", with: ".")) else {
            self.errorMessage = "Заполните все поля"
            return false
        }
        
        isLoading = true
        let amountInt = Int64(amountDouble * 100)
        
        let request = CreateTransReq(
            accountId: accId,
            amount: amountInt,
            categoryId: catId,
            comment: "",
            completedAt: Date(),
            isIncome: isIncome,
            name: name
        )
        
        do {
            let _: [String: String] = try await NetworkManager.shared.post(endpoint: "/transactions", body: request)
            return true
        } catch {
            self.errorMessage = error.localizedDescription
            isLoading = false
            return false
        }
    }
}
