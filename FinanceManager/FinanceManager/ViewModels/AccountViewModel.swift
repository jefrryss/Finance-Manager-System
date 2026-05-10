import Foundation
import Observation

@Observable
class AccountViewModel {
    var accounts: [Account] = []
    var isLoading = false
    
    // Вычисляем общий баланс в рублях
    var totalBalance: Double {
        let totalKopecks = accounts.reduce(0) { sum, account in
            sum + account.balance
        }
        return Double(totalKopecks) / 100.0
    }
    
    func fetchAccounts() async {
        isLoading = true
        do {
            self.accounts = try await NetworkManager.shared.fetch(endpoint: "/accounts")
            print("✅ Загружено счетов: \(accounts.count)")
        } catch {
            print("❌ Ошибка парсинга счетов: \(error)")
        }
        isLoading = false
    }
    
    func deleteAccount(_ account: Account) async -> Bool {
        do {
            try await NetworkManager.shared.delete(endpoint: "/accounts/\(account.accountId.uuidString.lowercased())")
            
            await fetchAccounts()
            return true
        } catch {
            print("❌ Ошибка при удалении счета: \(error)")
            return false
        }
    }
}
