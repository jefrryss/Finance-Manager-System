import Foundation

struct User: Codable, Identifiable {
    let userId: String
    let email: String
    let login: String
    let createdAt: Date
    let updatedAt: Date
    
    var id: String { userId }
}
