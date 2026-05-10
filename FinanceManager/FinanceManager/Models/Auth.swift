import Foundation

struct RegisterReq: Codable {
    let email: String
    let login: String
    let password: String
}

struct RegisterRes: Codable {
    let status: String
    let message: String
    let id: String
}