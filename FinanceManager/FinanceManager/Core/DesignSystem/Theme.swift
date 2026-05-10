import SwiftUI

enum AppTheme {
    static let bgPrimary = Color(hex: "#050505")
    static let bgSecondary = Color(hex: "#121212")
    static let accent = Color(hex: "#00E676")
    static let textPrimary = Color.white
    static let textSecondary = Color.white.opacity(0.6)
    
    static let cornerRadius: CGFloat = 16
    
    static var finexaBackground: some View {
        ZStack {
            bgPrimary.ignoresSafeArea()
            
            RadialGradient(
                colors: [accent.opacity(0.15), Color.clear],
                center: .bottomTrailing,
                startRadius: 0,
                endRadius: 500
            )
            .ignoresSafeArea()
            
            RadialGradient(
                colors: [accent.opacity(0.08), Color.clear],
                center: .topLeading,
                startRadius: 0,
                endRadius: 400
            )
            .ignoresSafeArea()
        }
    }
}

extension Color {
    init(hex: String) {
        var cleanHexCode = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        cleanHexCode = cleanHexCode.replacingOccurrences(of: "#", with: "")
        var rgb: UInt64 = 0
        Scanner(string: cleanHexCode).scanHexInt64(&rgb)
        let redValue = Double((rgb >> 16) & 0xFF) / 255.0
        let greenValue = Double((rgb >> 8) & 0xFF) / 255.0
        let blueValue = Double(rgb & 0xFF) / 255.0
        self.init(red: redValue, green: greenValue, blue: blueValue)
    }
}
