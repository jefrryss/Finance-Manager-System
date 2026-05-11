import SwiftUI

struct MainContainerView: View {
    @State private var selectedTab: Int = 0
    
    init() {
        let appearance = UITabBarAppearance()
        appearance.configureWithDefaultBackground()
        appearance.backgroundColor = UIColor(AppTheme.bgPrimary)
        appearance.shadowColor = .clear
        
        UITabBar.appearance().standardAppearance = appearance
        UITabBar.appearance().scrollEdgeAppearance = appearance
    }
    
    var body: some View {
        TabView(selection: $selectedTab) {
            HomeView()
                .tabItem {
                    Label("Главная", systemImage: "house.fill")
                }
                .tag(0)
            
            TransactionsView()
                .tabItem {
                    Label("Операции", systemImage: "arrow.left.arrow.right")
                }
                .tag(1)
            
            NavigationStack {
                Text("Раздел аналитики")
                    .navigationTitle("Анализ")
            }
            .tabItem {
                Label("Аналитика", systemImage: "chart.pie.fill")
            }
            .tag(2)
            
            ProfileView()
                .tabItem {
                    Label("Профиль", systemImage: "person.fill")
                }
                .tag(3)
        }
        .accentColor(AppTheme.accent)
    }
}
