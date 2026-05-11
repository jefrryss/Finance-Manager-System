import Foundation
import Observation

@Observable
class AddTransactionViewModel {
    var accounts: [Account] = []
    var categories: [TransactionCategory] = []
    
    var selectedAccountId: String?
    var selectedCategoryId: String?
    var name = ""
    var amountString = ""
    var isIncome = false
    
    var isCustomCategory = false
    var customCategoryName = ""
    
    var isLoading = false
    var errorMessage: String?
    
    func fetchFormOptions() async {
        do {
            async let fetchedAccounts: [Account] = try NetworkManager.shared.fetch(endpoint: "/accounts")
            async let fetchedCategories: [TransactionCategory] = try NetworkManager.shared.fetch(endpoint: "/categories")
            
            let (accs, cats) = try await (fetchedAccounts, fetchedCategories)
            
            await MainActor.run {
                self.accounts = accs
                self.categories = cats
                
                if let firstAcc = accs.first { self.selectedAccountId = firstAcc.accountId }
                
                if cats.isEmpty {
                    self.isCustomCategory = true
                } else {
                    updateCategorySelection()
                }
            }
        } catch {
            await MainActor.run {
                self.errorMessage = "Ошибка загрузки данных"
            }
        }
    }
    
    func updateCategorySelection() {
        let filteredCats = categories.filter { $0.isIncome == isIncome }
        selectedCategoryId = filteredCats.first?.categoryId
    }
    
    func saveTransaction() async -> Bool {
        guard let accId = selectedAccountId,
              let amountDouble = Double(amountString.replacingOccurrences(of: ",", with: ".")) else {
            self.errorMessage = "Заполните все поля"
            return false
        }
        
        if isCustomCategory && customCategoryName.trimmingCharacters(in: .whitespaces).isEmpty {
            self.errorMessage = "Введите название категории"
            return false
        }
        
        isLoading = true
        var finalCategoryId = selectedCategoryId
        
        if isCustomCategory {
            let newCatReq = CreateCategoryReq(
                name: customCategoryName.trimmingCharacters(in: .whitespaces),
                isIncome: isIncome,
                iconUrl: nil
            )
            
            do {
                let createdCat: CreateCategoryResponse = try await NetworkManager.shared.post(endpoint: "/categories", body: newCatReq)
                finalCategoryId = createdCat.categoryId
            } catch {
                print("Ошибка создания категории: \(error)")
                self.errorMessage = "Ошибка создания категории: \(error.localizedDescription)"
                isLoading = false
                return false
            }
        }
        
        guard let catId = finalCategoryId else {
            self.errorMessage = "Выберите категорию"
            isLoading = false
            return false
        }
        
        let amountInt = Int64(amountDouble * 100)
        
        let request = NewTransactionRequest(
            accountId: accId,
            categoryId: catId,
            name: name,
            isIncome: isIncome,
            amount: amountInt,
            completedAt: Date(),
            comment: "",
            currency: "RUB",
            bankFee: 0,
            status: "completed"
        )
        
        do {
            let _: [String: String] = try await NetworkManager.shared.post(endpoint: "/transactions", body: request)
            isLoading = false
            return true
        } catch {
            print("Ошибка создания транзакции: \(error)")
            self.errorMessage = error.localizedDescription
            isLoading = false
            return false
        }
    }
}
