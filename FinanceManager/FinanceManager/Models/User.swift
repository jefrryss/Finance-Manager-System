import Foundation

struct User: Codable, Identifiable {
    let userId: UUID
    let email: String
    let login: String
    let createdAt: Date
    let updatedAt: Date
    
    var id: UUID { userId }
}
