import SwiftUI

struct AddAccountView: View {
    @Environment(\.dismiss) var dismiss
    @State private var viewModel = AddAccountViewModel()
    var onSave: () -> Void
    
    let colors = ["#00E676", "#2196F3", "#FFEB3B", "#F44336", "#9C27B0"]
    
    var body: some View {
        NavigationStack {
            ZStack {
                AppTheme.finexaBackground
                VStack(spacing: 24) {
                    Text("Новый счет").font(.title.bold()).foregroundColor(.white)
                    
                    FintechTextField(icon: "creditcard", placeholder: "Название (напр. Сбер)", text: $viewModel.name)
                    FintechTextField(icon: "rublesign", placeholder: "Начальный баланс", text: $viewModel.balanceString).keyboardType(.numberPad)
                    
                    VStack(alignment: .leading) {
                        Text("Цвет счета").foregroundColor(AppTheme.textSecondary).font(.caption)
                        HStack {
                            ForEach(colors, id: \.self) { hex in
                                Circle()
                                    .fill(Color(hex: hex))
                                    .frame(width: 40, height: 40)
                                    .overlay(Circle().stroke(Color.white, lineWidth: viewModel.colorHex == hex ? 3 : 0))
                                    .onTapGesture { viewModel.colorHex = hex }
                            }
                        }
                    }
                    
                    Spacer()
                    
                    if let error = viewModel.errorMessage {
                        Text(error)
                            .foregroundColor(.red)
                            .font(.caption)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal)
                    }
                    
                    FintechButton(title: "Создать", isLoading: viewModel.isLoading, isDisabled: viewModel.name.isEmpty) {
                        Task {
                            if await viewModel.saveAccount() {
                                onSave()
                                dismiss()
                            }
                        }
                    }
                }
                .padding(24)
            }
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Отмена") { dismiss() }.foregroundColor(.white)
                }
            }
        }
    }
}
