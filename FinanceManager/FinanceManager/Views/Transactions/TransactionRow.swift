import SwiftUI

struct TransactionRow: View {
    let transaction: Transaction
    let category: TransactionCategory?
    
    var body: some View {
        HStack(spacing: 16) {
            ZStack {
                Circle()
                    .fill(AppTheme.bgSecondary)
                    .frame(width: 48, height: 48)
                
                Image(systemName: transaction.isIncome ? "arrow.down.left" : "arrow.up.right")
                    .foregroundColor(transaction.isIncome ? AppTheme.accent : .red)
                    .font(.system(size: 20, weight: .bold))
            }
            
            VStack(alignment: .leading, spacing: 4) {
                Text(transaction.name)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(AppTheme.textPrimary)
                    .lineLimit(1)
                
                HStack(spacing: 6) {
                    Text(transaction.completedAt.formatted(date: .abbreviated, time: .omitted))
                    
                    if let catName = category?.nameCategory {
                        Text("•")
                        Text(catName)
                            .lineLimit(1)
                    }
                }
                .font(.system(size: 13))
                .foregroundColor(AppTheme.textSecondary)
            }
            
            Spacer()

            Text("\(transaction.isIncome ? "+" : "-")\(Double(transaction.amount) / 100.0, specifier: "%.2f") ₽")
                .font(.system(size: 16, weight: .bold, design: .rounded))
                .foregroundColor(transaction.isIncome ? AppTheme.accent : AppTheme.textPrimary)
                .layoutPriority(1)
        }
        .padding(.vertical, 8)
    }
}
