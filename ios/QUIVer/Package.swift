// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "QUIVer",
    platforms: [
        .iOS(.v15)
    ],
    products: [
        .library(
            name: "QUIVer",
            targets: ["QUIVer"]),
    ],
    dependencies: [
        // libp2p-swift for P2P networking
        .package(url: "https://github.com/swift-libp2p/swift-libp2p.git", from: "0.1.0"),
        // Alamofire for HTTP networking
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.8.0"),
        // SwiftProtobuf for protocol buffers
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.25.0"),
    ],
    targets: [
        .target(
            name: "QUIVer",
            dependencies: [
                .product(name: "LibP2P", package: "swift-libp2p"),
                .product(name: "Alamofire", package: "Alamofire"),
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ]),
        .testTarget(
            name: "QUIVerTests",
            dependencies: ["QUIVer"]),
    ]
)