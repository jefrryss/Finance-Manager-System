import SwiftUI

struct TransactionRow: View {
    let transaction: Transaction
    
    var body: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(transaction.isIncome ? Color.green.opacity(0.15) : AppTheme.bgSecondary)
                    .frame(width: 50, height: 50)
                
                Image(systemName: transaction.isIncome ? "arrow.down.left" : "bag.fill")
                    .foregroundColor(transaction.isIncome ? .green : AppTheme.textPrimary)
                    .font(.system(size: 20, weight: .semibold))
            }
            
            VStack(alignment: .leading, spacing: 6) {
                Text(transaction.nameTransaction)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(AppTheme.textPrimary)
                
                Text(transaction.completedAt.formatted(date: .numeric, time: .shortened))
                    .font(.system(size: 13))
                    .foregroundColor(AppTheme.textSecondary)
            }
            
            Spacer()
            
            Text("\(transaction.isIncome ? "+" : "-") \(Double(transaction.amount) / 100.0, specifier: "%.2f") ₽")
                .font(.system(size: 16, weight: .bold, design: .rounded))
                .foregroundColor(transaction.isIncome ? .green : AppTheme.textPrimary)
        }
        .padding()
        .background(AppTheme.bgSecondary)
        .cornerRadius(AppTheme.cornerRadius)
    }
}
