import Foundation
import Observation

@Observable
class RegisterViewModel {
    var email = ""
    var login = ""
    var password = ""
    var isLoading = false
    var errorMessage: String?

    private let emailRegex = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"

    var isFormValid: Bool {
        let predicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
        return predicate.evaluate(with: email) && password.count >= 6 && !login.isEmpty
    }

    func register() async -> Bool {
        guard isFormValid else { return false }
        isLoading = true
        errorMessage = nil
        
        let req = RegisterReq(email: email, login: login, password: password)
        do {
            let res: RegisterRes = try await NetworkManager.shared.post(endpoint: "/users/register", body: req)
            NetworkManager.shared.currentUserId = res.id
            isLoading = false
            return true
        } catch {
            isLoading = false
            self.errorMessage = "Ошибка подключения или пользователь уже существует"
            return false
        }
    }
}
