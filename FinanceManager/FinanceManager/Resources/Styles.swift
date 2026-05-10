import SwiftUI

struct PremiumCardModifier: ViewModifier {
    var color: Color
    func body(content: Content) -> some View {
        content
            .padding(20)
            .background(
                ZStack {
                    RoundedRectangle(cornerRadius: 24, style: .continuous)
                        .fill(LinearGradient(
                            colors: [color, color.opacity(0.8), color.opacity(0.6)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ))
                    
                    Circle()
                        .fill(Color.white.opacity(0.1))
                        .frame(width: 150, height: 150)
                        .offset(x: 80, y: -40)
                }
            )
            .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
            .shadow(color: color.opacity(0.3), radius: 15, x: 0, y: 10)
    }
}

extension View {
    func premiumCard(color: Color = .blue) -> some View {
        modifier(PremiumCardModifier(color: color))
    }
}
