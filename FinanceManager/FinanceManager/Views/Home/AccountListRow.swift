import SwiftUI

struct AccountListRow: View {
    let account: Account
    
    var body: some View {
        HStack(spacing: 16) {
            ZStack {
                RoundedRectangle(cornerRadius: 12)
                    .fill(Color(hex: account.colorHex).opacity(0.2))
                    .frame(width: 48, height: 48)
                
                Image(systemName: account.isImported ? "building.columns.fill" : "wallet.pass.fill")
                    .foregroundColor(Color(hex: account.colorHex))
                    .font(.system(size: 20))
            }
            
            VStack(alignment: .leading, spacing: 4) {
                Text(account.nameAccount)
                    .font(.system(size: 17, weight: .semibold))
                    .foregroundColor(AppTheme.textPrimary)
                
                Text(account.isImported ? "Т-Банк" : "Личный счет")
                    .font(.system(size: 13))
                    .foregroundColor(AppTheme.textSecondary)
            }
            
            Spacer()
            
            VStack(alignment: .trailing, spacing: 4) {
                Text("\(Double(account.balance) / 100.0, specifier: "%.1f") ₽")
                    .font(.system(size: 17, weight: .bold, design: .rounded))
                    .foregroundColor(AppTheme.textPrimary)
                
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundColor(AppTheme.textSecondary)
            }
        }
        .padding()
        .background(AppTheme.bgSecondary.opacity(0.5))
        .cornerRadius(20)
    }
}
