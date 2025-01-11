## Services Overview

### 1. **Extractor Service**
- **Purpose**: Extract comments from the input YouTube Short.
- **Implementation**: Runs entirely on the client side to reduce costs and minimize the risk of being blocked.

---

### 2. **Processing Service**
- **Purpose**: Process the extracted comments to identify relevant hits.
- **Key Features**:
  - **Regex Matching**:
    - Use regular expressions to shortlist potential hits from the extracted comments.
  - **LLM Integration** (Optional):
    - Explore the feasibility of using a Language Learning Model (LLM) to improve hit detection accuracy and handle edge cases.

---

### 3. **Movie Database**
- **Purpose**: Maintain a comprehensive list of known movies and series.
- **Usage**:
  - Cross-reference shortlisted comments with the database to increase confidence in results.
- **Implementation**:
  - This service can be hosted separately (server-side) since it doesn't interact directly with sensitive client data.

---

### 4. **Frontend Application**
- **Purpose**: Provide a simple and intuitive Single Page Application (SPA) interface.
- **Functionality**:
  - Allow users to input YouTube Shorts links.
  - Display results, including relevant hits.

---

### Notes
- **Client-Side Execution**:
  - All services, except the `Movie Database`, should run on the client side.
  - **Benefits**:
    - Reduces operational costs.
    - Mitigates the risk of the `Extractor Service` being blocked due to repeated server-side requests.
