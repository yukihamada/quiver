import SwiftUI

@main
struct QUIVerApp: App {
    @StateObject private var networkManager = P2PNetworkManager()
    @StateObject private var modelManager = ModelManager()
    
    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(networkManager)
                .environmentObject(modelManager)
                .onAppear {
                    networkManager.connect()
                }
        }
    }
}

// P2P Network Manager
class P2PNetworkManager: ObservableObject {
    @Published var isConnected = false
    @Published var availableProviders: [Provider] = []
    @Published var currentLatency: Double = 0
    
    private let gatewayURL = "https://api.quiver.network"
    private var webSocketTask: URLSessionWebSocketTask?
    
    func connect() {
        // Use WebSocket for real-time P2P-like communication
        guard let url = URL(string: "wss://gateway.quiver.network/ws") else { return }
        
        let session = URLSession(configuration: .default)
        webSocketTask = session.webSocketTask(with: url)
        webSocketTask?.resume()
        
        receiveMessage()
        isConnected = true
        
        // Discover providers
        discoverProviders()
    }
    
    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            switch result {
            case .success(let message):
                switch message {
                case .string(let text):
                    self?.handleMessage(text)
                case .data(let data):
                    self?.handleData(data)
                @unknown default:
                    break
                }
                self?.receiveMessage()
            case .failure(let error):
                print("WebSocket error: \(error)")
                self?.isConnected = false
            }
        }
    }
    
    private func handleMessage(_ message: String) {
        // Parse P2P messages
        if let data = message.data(using: .utf8),
           let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
            
            if let type = json["type"] as? String {
                switch type {
                case "provider_update":
                    updateProviders(json)
                case "latency":
                    if let latency = json["latency"] as? Double {
                        DispatchQueue.main.async {
                            self.currentLatency = latency
                        }
                    }
                default:
                    break
                }
            }
        }
    }
    
    private func handleData(_ data: Data) {
        // Handle binary protocol messages
    }
    
    private func updateProviders(_ json: [String: Any]) {
        if let providers = json["providers"] as? [[String: Any]] {
            DispatchQueue.main.async {
                self.availableProviders = providers.compactMap { Provider(json: $0) }
            }
        }
    }
    
    func discoverProviders() {
        // Request provider list
        let message = URLSessionWebSocketTask.Message.string("""
        {"type": "discover_providers"}
        """)
        webSocketTask?.send(message) { error in
            if let error = error {
                print("Send error: \(error)")
            }
        }
    }
    
    func sendInferenceRequest(_ request: InferenceRequest, to provider: Provider) async throws -> InferenceResponse {
        // Select optimal provider based on model availability and latency
        let encoder = JSONEncoder()
        let requestData = try encoder.encode(request)
        
        let message = URLSessionWebSocketTask.Message.data(requestData)
        
        return try await withCheckedThrowingContinuation { continuation in
            webSocketTask?.send(message) { error in
                if let error = error {
                    continuation.resume(throwing: error)
                }
            }
            
            // Wait for response
            // In real implementation, would use request ID for matching
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                let response = InferenceResponse(
                    completion: "QUIVer P2Pネットワークからの応答",
                    model: request.model,
                    latency: self.currentLatency,
                    providerId: provider.id
                )
                continuation.resume(returning: response)
            }
        }
    }
}

// Model Manager
class ModelManager: ObservableObject {
    @Published var availableModels: [AIModel] = []
    @Published var popularModels: [AIModel] = []
    @Published var downloadedModels: Set<String> = []
    
    init() {
        loadModels()
    }
    
    private func loadModels() {
        // Initialize with all supported models
        availableModels = [
            AIModel(id: "qwen3:0.6b", name: "Qwen3 0.6B", size: "600MB", category: .general),
            AIModel(id: "qwen3:3b", name: "Qwen3 3B", size: "3GB", category: .general),
            AIModel(id: "qwen3:7b", name: "Qwen3 7B", size: "7GB", category: .general),
            AIModel(id: "qwen3:14b", name: "Qwen3 14B", size: "14GB", category: .general),
            AIModel(id: "qwen3:32b", name: "Qwen3 32B", size: "32GB", category: .general),
            AIModel(id: "qwen3-coder:30b", name: "Qwen3 Coder", size: "30GB", category: .coding),
            AIModel(id: "gpt-oss:20b", name: "GPT-OSS 20B", size: "20GB", category: .general),
            AIModel(id: "jan-nano:32k", name: "Jan-Nano 32K", size: "8GB", category: .longContext),
            AIModel(id: "jan-nano:128k", name: "Jan-Nano 128K", size: "16GB", category: .longContext),
            AIModel(id: "llama3.2:3b", name: "Llama 3.2", size: "3GB", category: .general),
            AIModel(id: "mistral:7b", name: "Mistral 7B", size: "7GB", category: .general),
        ]
        
        // Popular models based on network demand
        updatePopularModels()
    }
    
    func updatePopularModels() {
        // In real app, this would fetch from network
        popularModels = Array(availableModels.prefix(5))
    }
    
    func isModelAvailable(_ modelId: String) -> Bool {
        // Check if any connected provider has this model
        return true // Simplified for demo
    }
}

// Data Models
struct Provider: Identifiable {
    let id: String
    let peerID: String
    let models: [String]
    let latency: Double
    let location: String
    
    init?(json: [String: Any]) {
        guard let id = json["id"] as? String,
              let peerID = json["peer_id"] as? String,
              let models = json["models"] as? [String],
              let latency = json["latency"] as? Double else {
            return nil
        }
        
        self.id = id
        self.peerID = peerID
        self.models = models
        self.latency = latency
        self.location = json["location"] as? String ?? "Unknown"
    }
}

struct AIModel: Identifiable {
    let id: String
    let name: String
    let size: String
    let category: ModelCategory
    
    enum ModelCategory {
        case general
        case coding
        case longContext
    }
}

struct InferenceRequest: Codable {
    let prompt: String
    let model: String
    let maxTokens: Int
    let temperature: Double
}

struct InferenceResponse: Codable {
    let completion: String
    let model: String
    let latency: Double
    let providerId: String
}