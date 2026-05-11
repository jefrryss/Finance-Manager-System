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
            let _: [String: String] = try await NetworkManager.shared.put(endpoint: "/users/change_password", body: ["password": newPassword])
            print("✅ Пароль успешно изменён")
            self.errorMessage = nil
            return true
        } catch {
            print("❌ Ошибка смены пароля: \(error)")
            print("Описание: \(error.localizedDescription)")
            self.errorMessage = error.localizedDescription.isEmpty ? "Не удалось сменить пароль" : error.localizedDescription
            return false
        }
    }
}
