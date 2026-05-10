import SwiftUI

@main
struct FinanceManagerApp: App {
    @AppStorage("isLoggedIn") private var isLoggedIn = false

    var body: some Scene {
        WindowGroup {
            if isLoggedIn {
                MainContainerView()
                    .transition(.opacity)
            } else {
                WelcomeView(isLoggedIn: $isLoggedIn)
            }
        }
    }
}
