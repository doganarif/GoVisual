flowchart TD
subgraph "Developer Workflow"
A1[Run Local App with GoVisual] --> A2["Open Browser: localhost:8080/__viz"]
A2 --> A3[Monitor Requests in Dashboard]
A3 --> A4[Debug Issues & Optimize App]
end

    subgraph "HTTP Request Flow"
        B1[Local API Request] --> B2[govisual.Wrap]
        B2 --> B3{Is Dashboard Path?}
        B3 -->|Yes| B4[Serve GoVisual Dashboard]
        B3 -->|No| B5[Process Through Middleware]
        B5 --> B6[Forward to App Handler]
        B6 --> B7[Return to User]
    end
    
    subgraph "Visualization & Debug Features"
        C1[Real-time Request Table] --> C2[Filter & Search]
        C1 --> C3[Detailed Request View]
        C3 --> C4[Headers Inspection]
        C3 --> C5[Request Body Viewer]
        C3 --> C6[Response Body Viewer]
        C3 --> C7[Status & Timing Info]
        C1 --> C8[Middleware Trace Visualization]
        C8 --> C9[Performance Bottleneck Detection]
    end
    
    subgraph "Data Capture System"
        D1[Intercept HTTP Request] --> D2[Capture Headers]
        D1 --> D3[Capture Request Body]
        D1 --> D4[Measure Start Time]
        D5[Wrap Response Writer] --> D6[Capture Status Code]
        D5 --> D7[Capture Response Body]
        D5 --> D8[Calculate Duration]
        D9[Track Middleware Flow] --> D10[Record Execution Times]
    end
    
    subgraph "Development Configuration"
        E1[Feature Flags] --> E2[Enable in Dev Only]
        E1 --> E3[Disable in Production]
        E4[Logging Options] --> E5[Request Body Capture]
        E4 --> E6[Response Body Capture]
        E7[Storage Options] --> E8[Memory Limits]
    end
    
    B7 --> A3
    D10 --> C8
    D8 --> C7
    D3 --> C5
    D7 --> C6
    E2 --> A1