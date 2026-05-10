import Foundation
import Observation

struct CreateAccountRequest: Codable {
    let name: String
    let initialBalance: Int64
    let currency: String
    let colorHex: String
    let accountType: String
    let isImported: Bool
}

struct CreateAccountResponse: Codable {
    let status: String
    let message: String?
}

@Observable
class AddAccountViewModel {
    var name = ""
    var balanceString = ""
    var currency = "RUB"
    var colorHex = "#00E676"
    var accountType = "manual"
    
    var isLoading = false
    var errorMessage: String?
    
    func saveAccount() async -> Bool {
        guard !name.isEmpty else {
            errorMessage = "Введите название счета"
            return false
        }
        
        isLoading = true
        errorMessage = nil
        
        let initialBalanceValue = (Int64(balanceString) ?? 0) * 100
        
        let requestBody = CreateAccountRequest(
            name: name,
            initialBalance: initialBalanceValue,
            currency: currency,
            colorHex: colorHex,
            accountType: accountType,
            isImported: false
        )
        
        do {
            let res: CreateAccountResponse = try await NetworkManager.shared.post(endpoint: "/accounts", body: requestBody)
            print("✅ Счет создан: \(res.message ?? "")")
            isLoading = false
            return true
        } catch {
            isLoading = false
            self.errorMessage = "Ошибка сервера: \(error.localizedDescription)"
            print("❌ Ошибка создания счета: \(error)")
            return false
        }
    }
}
