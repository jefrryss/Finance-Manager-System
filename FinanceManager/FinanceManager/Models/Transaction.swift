import Foundation

struct Transaction: Codable, Identifiable {
    var id: String { transactionId }
    let transactionId: String
    let accountId: String
    let categoryId: String?
    let nameTransaction: String
    let amount: Int64
    let isIncome: Bool
    let completedAt: Date
    let comment: String?
    let currency: String
    let bankFee: Int64?
    let isImported: Bool?
    let externalTransactionId: String?
    let isHidden: Bool?
    
    var name: String { nameTransaction }
    
}

struct NewTransactionRequest: Codable {
    let accountId: String
    let categoryId: String
    let name: String
    let isIncome: Bool
    let amount: Int64
    let completedAt: Date
    let comment: String?
    let currency: String
    let bankFee: Int64
    let status: String
} 
