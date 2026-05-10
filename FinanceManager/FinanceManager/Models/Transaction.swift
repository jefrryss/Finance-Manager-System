import Foundation

struct Transaction: Codable, Identifiable {
    let transactionId: UUID
    let userId: UUID
    let accountId: UUID
    let categoryId: UUID?
    let nameTransaction: String
    let isIncome: Bool
    let amount: Int64
    let completedAt: Date
    let isHidden: Bool
    let isImported: Bool
    let comment: String?
    
    var id: UUID { transactionId }
}

struct CreateTransReq: Codable {
    let accountId: UUID
    let amount: Int64
    let categoryId: UUID
    let comment: String
    let completedAt: Date
    let isIncome: Bool
    let name: String
}
