import SwiftUI

struct ContentView: View {
    @EnvironmentObject var networkManager: P2PNetworkManager
    @EnvironmentObject var modelManager: ModelManager
    
    @State private var selectedTab = 0
    
    var body: some View {
        TabView(selection: $selectedTab) {
            ChatView()
                .tabItem {
                    Label("チャット", systemImage: "message.fill")
                }
                .tag(0)
            
            ModelsView()
                .tabItem {
                    Label("モデル", systemImage: "cube.box.fill")
                }
                .tag(1)
            
            NetworkView()
                .tabItem {
                    Label("ネットワーク", systemImage: "network")
                }
                .tag(2)
            
            SettingsView()
                .tabItem {
                    Label("設定", systemImage: "gear")
                }
                .tag(3)
        }
        .accentColor(.purple)
    }
}

// Chat View
struct ChatView: View {
    @EnvironmentObject var networkManager: P2PNetworkManager
    @EnvironmentObject var modelManager: ModelManager
    
    @State private var messages: [ChatMessage] = []
    @State private var inputText = ""
    @State private var selectedModel = "llama3.2:3b"
    @State private var isLoading = false
    
    var body: some View {
        NavigationView {
            VStack {
                // Model selector
                HStack {
                    Menu {
                        ForEach(modelManager.availableModels) { model in
                            Button(model.name) {
                                selectedModel = model.id
                            }
                        }
                    } label: {
                        HStack {
                            Image(systemName: "cube.box")
                            Text(modelName(for: selectedModel))
                            Image(systemName: "chevron.down")
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 8)
                        .background(Color.purple.opacity(0.1))
                        .cornerRadius(8)
                    }
                    
                    Spacer()
                    
                    if networkManager.isConnected {
                        HStack {
                            Circle()
                                .fill(Color.green)
                                .frame(width: 8, height: 8)
                            Text("\(Int(networkManager.currentLatency))ms")
                                .font(.caption)
                        }
                    }
                }
                .padding()
                
                // Chat messages
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 12) {
                            ForEach(messages) { message in
                                ChatBubble(message: message)
                                    .id(message.id)
                            }
                            
                            if isLoading {
                                HStack {
                                    ProgressView()
                                        .progressViewStyle(CircularProgressViewStyle())
                                    Text("生成中...")
                                        .foregroundColor(.gray)
                                }
                                .padding()
                            }
                        }
                        .padding()
                    }
                    .onChange(of: messages.count) { _ in
                        withAnimation {
                            proxy.scrollTo(messages.last?.id, anchor: .bottom)
                        }
                    }
                }
                
                // Input field
                HStack {
                    TextField("メッセージを入力", text: $inputText)
                        .textFieldStyle(RoundedBorderTextFieldStyle())
                        .onSubmit {
                            sendMessage()
                        }
                    
                    Button(action: sendMessage) {
                        Image(systemName: "paperplane.fill")
                            .foregroundColor(.white)
                            .padding(10)
                            .background(Color.purple)
                            .clipShape(Circle())
                    }
                    .disabled(inputText.isEmpty || isLoading)
                }
                .padding()
            }
            .navigationTitle("QUIVer AI")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
    
    private func modelName(for id: String) -> String {
        modelManager.availableModels.first { $0.id == id }?.name ?? id
    }
    
    private func sendMessage() {
        guard !inputText.isEmpty else { return }
        
        let userMessage = ChatMessage(
            id: UUID(),
            text: inputText,
            isUser: true,
            timestamp: Date()
        )
        messages.append(userMessage)
        
        let prompt = inputText
        inputText = ""
        isLoading = true
        
        Task {
            do {
                // Find best provider for selected model
                let provider = networkManager.availableProviders.first ?? 
                    Provider(json: ["id": "default", "peer_id": "local", "models": [selectedModel], "latency": 10])!
                
                let request = InferenceRequest(
                    prompt: prompt,
                    model: selectedModel,
                    maxTokens: 1000,
                    temperature: 0.7
                )
                
                let response = try await networkManager.sendInferenceRequest(request, to: provider)
                
                await MainActor.run {
                    let aiMessage = ChatMessage(
                        id: UUID(),
                        text: response.completion,
                        isUser: false,
                        timestamp: Date(),
                        model: response.model,
                        latency: response.latency
                    )
                    messages.append(aiMessage)
                    isLoading = false
                }
            } catch {
                await MainActor.run {
                    let errorMessage = ChatMessage(
                        id: UUID(),
                        text: "エラーが発生しました: \(error.localizedDescription)",
                        isUser: false,
                        timestamp: Date()
                    )
                    messages.append(errorMessage)
                    isLoading = false
                }
            }
        }
    }
}

// Chat Bubble
struct ChatBubble: View {
    let message: ChatMessage
    
    var body: some View {
        HStack {
            if message.isUser { Spacer() }
            
            VStack(alignment: message.isUser ? .trailing : .leading, spacing: 4) {
                Text(message.text)
                    .padding(12)
                    .background(message.isUser ? Color.purple : Color.gray.opacity(0.2))
                    .foregroundColor(message.isUser ? .white : .primary)
                    .cornerRadius(16)
                
                HStack(spacing: 8) {
                    if let model = message.model {
                        Text(model)
                            .font(.caption2)
                    }
                    
                    if let latency = message.latency {
                        Text("\(Int(latency))ms")
                            .font(.caption2)
                    }
                    
                    Text(message.timestamp, style: .time)
                        .font(.caption2)
                }
                .foregroundColor(.gray)
            }
            
            if !message.isUser { Spacer() }
        }
    }
}

// Models View
struct ModelsView: View {
    @EnvironmentObject var modelManager: ModelManager
    @State private var searchText = ""
    
    var filteredModels: [AIModel] {
        if searchText.isEmpty {
            return modelManager.availableModels
        } else {
            return modelManager.availableModels.filter { 
                $0.name.localizedCaseInsensitiveContains(searchText) ||
                $0.id.localizedCaseInsensitiveContains(searchText)
            }
        }
    }
    
    var body: some View {
        NavigationView {
            List {
                // Popular models section
                Section("人気のモデル") {
                    ForEach(modelManager.popularModels) { model in
                        ModelRow(model: model)
                    }
                }
                
                // All models section
                Section("すべてのモデル") {
                    ForEach(filteredModels) { model in
                        ModelRow(model: model)
                    }
                }
            }
            .searchable(text: $searchText, prompt: "モデルを検索")
            .navigationTitle("AIモデル")
            .refreshable {
                modelManager.updatePopularModels()
            }
        }
    }
}

// Model Row
struct ModelRow: View {
    let model: AIModel
    @EnvironmentObject var modelManager: ModelManager
    
    var isAvailable: Bool {
        modelManager.isModelAvailable(model.id)
    }
    
    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(model.name)
                    .font(.headline)
                
                HStack {
                    Label(model.size, systemImage: "internaldrive")
                    
                    if isAvailable {
                        Label("利用可能", systemImage: "checkmark.circle.fill")
                            .foregroundColor(.green)
                    }
                }
                .font(.caption)
                .foregroundColor(.secondary)
            }
            
            Spacer()
            
            Image(systemName: categoryIcon(for: model.category))
                .foregroundColor(.purple)
        }
        .padding(.vertical, 4)
    }
    
    private func categoryIcon(for category: AIModel.ModelCategory) -> String {
        switch category {
        case .general:
            return "cube.box"
        case .coding:
            return "chevron.left.forwardslash.chevron.right"
        case .longContext:
            return "doc.text"
        }
    }
}

// Network View
struct NetworkView: View {
    @EnvironmentObject var networkManager: P2PNetworkManager
    
    var body: some View {
        NavigationView {
            List {
                // Connection status
                Section("接続状態") {
                    HStack {
                        Text("ステータス")
                        Spacer()
                        if networkManager.isConnected {
                            Label("接続済み", systemImage: "wifi")
                                .foregroundColor(.green)
                        } else {
                            Label("未接続", systemImage: "wifi.slash")
                                .foregroundColor(.red)
                        }
                    }
                    
                    HStack {
                        Text("レイテンシ")
                        Spacer()
                        Text("\(Int(networkManager.currentLatency))ms")
                            .foregroundColor(.secondary)
                    }
                }
                
                // Available providers
                Section("利用可能なプロバイダー (\(networkManager.availableProviders.count))") {
                    ForEach(networkManager.availableProviders) { provider in
                        VStack(alignment: .leading, spacing: 4) {
                            Text(provider.location)
                                .font(.headline)
                            
                            HStack {
                                Label("\(Int(provider.latency))ms", systemImage: "speedometer")
                                
                                Label("\(provider.models.count)モデル", systemImage: "cube.box")
                            }
                            .font(.caption)
                            .foregroundColor(.secondary)
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
            .navigationTitle("P2Pネットワーク")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("更新") {
                        networkManager.discoverProviders()
                    }
                }
            }
        }
    }
}

// Settings View
struct SettingsView: View {
    @AppStorage("preferredLatency") private var preferredLatency = 100
    @AppStorage("autoSelectModel") private var autoSelectModel = true
    @AppStorage("cacheResponses") private var cacheResponses = true
    
    var body: some View {
        NavigationView {
            Form {
                Section("パフォーマンス") {
                    VStack(alignment: .leading) {
                        Text("最大許容レイテンシ: \(preferredLatency)ms")
                        Slider(value: Binding(
                            get: { Double(preferredLatency) },
                            set: { preferredLatency = Int($0) }
                        ), in: 10...500, step: 10)
                    }
                    
                    Toggle("モデルを自動選択", isOn: $autoSelectModel)
                    
                    Toggle("レスポンスをキャッシュ", isOn: $cacheResponses)
                }
                
                Section("プライバシー") {
                    NavigationLink("プライバシー設定") {
                        PrivacySettingsView()
                    }
                }
                
                Section("情報") {
                    HStack {
                        Text("バージョン")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }
                    
                    Link("QUIVer Network", destination: URL(string: "https://quiver.network")!)
                    
                    Link("ソースコード", destination: URL(string: "https://github.com/yukihamada/quiver")!)
                }
            }
            .navigationTitle("設定")
        }
    }
}

// Privacy Settings
struct PrivacySettingsView: View {
    @AppStorage("shareAnalytics") private var shareAnalytics = false
    @AppStorage("localProcessing") private var localProcessing = true
    
    var body: some View {
        Form {
            Section {
                Toggle("分析データを共有", isOn: $shareAnalytics)
                
                Toggle("可能な限りローカル処理", isOn: $localProcessing)
            }
            
            Section {
                Text("QUIVerは分散型ネットワークを使用していますが、あなたのプライバシーを重視しています。")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
        .navigationTitle("プライバシー設定")
        .navigationBarTitleDisplayMode(.inline)
    }
}

// Data Models
struct ChatMessage: Identifiable {
    let id: UUID
    let text: String
    let isUser: Bool
    let timestamp: Date
    var model: String? = nil
    var latency: Double? = nil
}

// Preview
struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
            .environmentObject(P2PNetworkManager())
            .environmentObject(ModelManager())
    }
}