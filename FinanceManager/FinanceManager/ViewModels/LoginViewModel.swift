import Foundation
import Observation

struct LoginRequest: Codable {
    let identifier: String
    let password: String
}

struct LoginResponse: Codable {
    let status: String
    let token: String
}

@Observable
class LoginViewModel {
    var identifier = ""
    var password = ""
    var isLoading = false
    var errorMessage: String?
    
    func loginUser() async -> Bool {
        guard !identifier.isEmpty, !password.isEmpty else {
            errorMessage = "Заполни все поля"
            return false
        }
        
        isLoading = true
        errorMessage = nil
        
        let requestBody = LoginRequest(identifier: identifier, password: password)
        
        do {
            let res: LoginResponse = try await NetworkManager.shared.post(endpoint: "/users/login", body: requestBody)
            
            NetworkManager.shared.currentUserId = res.token
            
            isLoading = false
            return true
        } catch {
            isLoading = false
            self.errorMessage = "Неверный логин/почта или пароль"
            print("❌ Ошибка входа: \(error)")
            return false
        }
    }
}
