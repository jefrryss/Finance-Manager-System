import Foundation
import Observation

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
            let _: CreateAccountResponse = try await NetworkManager.shared.post(endpoint: "/accounts", body: requestBody)
            isLoading = false
            return true
        } catch {
            isLoading = false
            self.errorMessage = "Ошибка сервера: \(error.localizedDescription)"
            return false
        }
    }
    
    func importFromPDF(fileURL: URL) async -> Bool {
        isLoading = true
        errorMessage = nil
        
        guard fileURL.startAccessingSecurityScopedResource() else {
            errorMessage = "Нет доступа к файлу"
            isLoading = false
            return false
        }
        
        defer { fileURL.stopAccessingSecurityScopedResource() }
        
        do {
            let pdfData = try Data(contentsOf: fileURL)
            let _: ImportAccountResponse = try await NetworkManager.shared.uploadPDF(
                endpoint: "/accounts/import/pdf",
                fileData: pdfData,
                accountName: name
            )
            
            isLoading = false
            return true
        } catch {
            isLoading = false
            self.errorMessage = "Ошибка импорта: \(error.localizedDescription)"
            return false
        }
    }
}
