import Foundation
import Observation
import SwiftUI

@Observable
class ProfileViewModel {
    var isLoading = false
    var errorMessage: String?
    
    func logout() {
        NetworkManager.shared.currentUserId = nil

        UserDefaults.standard.set(false, forKey: "isLoggedIn")
    }
    
    func changePassword(newPassword: String) async -> Bool {
        isLoading = true
        defer { isLoading = false }
        
        do {
            let _: [String: String] = try await NetworkManager.shared.post(endpoint: "/users/change_password", body: ["password": newPassword])
            return true
        } catch {
            self.errorMessage = "Не удалось сменить пароль"
            return false
        }
    }
}
