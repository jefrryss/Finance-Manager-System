import Foundation
import Observation

@Observable
class TransactionViewModel {
    var transactions: [Transaction] = []
    var isLoading = false
    var errorMessage: String?
    
    func fetchTransactions() async {
        isLoading = true
        errorMessage = nil
        do {
            self.transactions = try await NetworkManager.shared.fetch(endpoint: "/transactions")
        } catch {
            self.errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}
