import Foundation

struct TransactionCategory: Codable, Identifiable {
    let categoryId: UUID
    let userId: UUID
    let nameCategory: String
    let isIncome: Bool
    let isCustom: Bool
    let iconUrl: String?
    
    var id: UUID { categoryId }
}
