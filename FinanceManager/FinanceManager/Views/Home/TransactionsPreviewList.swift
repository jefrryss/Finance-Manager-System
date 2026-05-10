import SwiftUI

struct TransactionsPreviewList: View {
    let transactions: [Transaction]
    
    var body: some View {
        if transactions.isEmpty {
            Text("Пока нет операций")
                .foregroundColor(AppTheme.textSecondary)
                .frame(maxWidth: .infinity)
                .padding()
        } else {
            VStack(spacing: 12) {
                ForEach(transactions.prefix(3)) { item in // Берем только 3 последние
                    HStack(spacing: 16) {
                        ZStack {
                            Circle()
                                .fill(item.isIncome ? AppTheme.accent.opacity(0.1) : AppTheme.bgSecondary)
                                .frame(width: 48, height: 48)
                            Image(systemName: item.isIncome ? "arrow.down.left" : "cart.fill")
                                .foregroundColor(item.isIncome ? AppTheme.accent : AppTheme.textPrimary)
                        }
                        
                        VStack(alignment: .leading, spacing: 4) {
                            Text(item.nameTransaction)
                                .font(.system(size: 16, weight: .semibold))
                                .foregroundColor(AppTheme.textPrimary)
                            Text(item.completedAt.formatted(date: .omitted, time: .shortened))
                                .font(.system(size: 13))
                                .foregroundColor(AppTheme.textSecondary)
                        }
                        
                        Spacer()
                        
                        Text("\(item.isIncome ? "+" : "-")\(String(format: "%.0f", Double(item.amount)/100.0)) ₽")
                            .font(.system(size: 16, weight: .bold, design: .rounded))
                            .foregroundColor(item.isIncome ? AppTheme.accent : AppTheme.textPrimary)
                    }
                    .padding()
                    .background(AppTheme.bgSecondary.opacity(0.5))
                    .cornerRadius(AppTheme.cornerRadius)
                    .padding(.horizontal, 24)
                }
            }
        }
    }
}
