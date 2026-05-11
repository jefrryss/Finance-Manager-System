import Foundation

struct Account: Codable, Identifiable {
    var id: String { accountId }
    let accountId: String
    let userId: String?
    let nameAccount: String
    let accountType: String
    let currency: String
    var balance: Int64
    let colorHex: String
    let isImported: Bool
    let isArchived: Bool?
    let externalAccountId: String?
    let createdAt: Date?
    let updatedAt: Date?
    let lastSyncedAt: Date?
}

struct CreateAccountRequest: Codable {
    let name: String
    let initialBalance: Int64
    let currency: String
    let colorHex: String
    let accountType: String
    let isImported: Bool
}

struct CreateAccountResponse: Codable {
    let status: String
    let message: String?
}

struct ImportAccountResponse: Codable {
    let status: String
    let accountId: String
    let importedTransactions: Int
    let balance: Int64
}
