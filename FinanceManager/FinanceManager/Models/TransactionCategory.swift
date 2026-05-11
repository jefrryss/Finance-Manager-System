import Foundation

struct TransactionCategory: Codable, Identifiable {
    var id: String { categoryId }
    let categoryId: String
    let nameCategory: String
    let isIncome: Bool
    let isCustom: Bool?
    let iconUrl: String?
    let userId: String?
}

struct CreateCategoryReq: Codable {
    let name: String
    let isIncome: Bool
    let iconUrl: String?
}

struct CreateCategoryResponse: Codable {
    let status: String
    let categoryId: String
} 
