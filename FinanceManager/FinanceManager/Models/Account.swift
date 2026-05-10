import Foundation

struct Account: Codable, Identifiable {
    let accountId: UUID
    let userId: UUID
    var balance: Int64
    let isImported: Bool
    let externalAccountId: String?
    let accountType: String
    let colorHex: String
    let isArchived: Bool
    let nameAccount: String
    let currency: String
    let lastSyncedAt: Date?
    let createdAt: Date
    
    var id: UUID { accountId }
}
