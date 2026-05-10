import SwiftUI

struct FintechTextField: View {
    var icon: String
    var placeholder: String
    @Binding var text: String
    var isSecure: Bool = false
    
    var body: some View {
        HStack(spacing: 15) {
            Image(systemName: icon)
                .foregroundColor(AppTheme.accent)
                .frame(width: 20)
            
            ZStack(alignment: .leading) {
                if text.isEmpty {
                    Text(placeholder)
                        .foregroundColor(AppTheme.textSecondary)
                }
                
                if isSecure {
                    SecureField("", text: $text)
                        .foregroundColor(.white)
                        .tint(AppTheme.accent)
                } else {
                    TextField("", text: $text)
                        .foregroundColor(.white)
                        .tint(AppTheme.accent)
                }
            }
        }
        .padding()
        .background(AppTheme.bgSecondary)
        .cornerRadius(AppTheme.cornerRadius)
        .overlay(
            RoundedRectangle(cornerRadius: AppTheme.cornerRadius)
                .stroke(Color.white.opacity(0.1), lineWidth: 1)
        )
    }
}


struct FintechButton: View {
    var title: String
    var isLoading: Bool
    var isDisabled: Bool
    var action: () -> Void
    
    var body: some View {
        Button {
            action()
        } label: {
            ZStack {
                if isLoading {
                    ProgressView().tint(AppTheme.bgPrimary)
                } else {
                    Text(title)
                        .font(.system(size: 18, weight: .bold))
                }
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(isDisabled ? Color.gray.opacity(0.3) : AppTheme.accent)
            .foregroundColor(AppTheme.bgPrimary)
            .cornerRadius(AppTheme.cornerRadius)
        }
        .disabled(isDisabled || isLoading)
    }
}

struct StubButton: View {
    var title: String
    
    var body: some View {
        Text(title)
            .font(.system(size: 18, weight: .semibold))
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .foregroundColor(AppTheme.textPrimary)
            .overlay(
                RoundedRectangle(cornerRadius: AppTheme.cornerRadius)
                    .stroke(Color.white.opacity(0.2), lineWidth: 1)
            )
            .contentShape(Rectangle())
    }
}
