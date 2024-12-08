
# CNAD - Car Sharing System (Casey Tan Wee Liang)

**INTRODUCTION**  
In an era marked by sustainable transportation and shared economies, electric carsharing platforms have emerged as a cornerstone of modern urban mobility. This project aims to design and implement a fully functional electric car-sharing system using Go, with features catering to diverse user needs and real-world application scenarios. With an emphasis on practical and scalable solutions, the system includes user membership tiers, promotional discounts, and an accurate billing mechanism. 


## System Architecture

The system consists of three core microservices:

**User Service (Port: 8080)**  
Handles user authentication, profile management, and membership tiers.
- Authentication and authorization
- User profile management
- JWT token generation and validation
- Membership tier management

**Vehicle Service (Port: 8085)**   
Manages vehicle fleet, bookings, and real-time vehicle status.
- Vehicle inventory management
- Booking creation and management
- Real-time vehicle status tracking
- Availability monitoring

**Billing Service (Port: 8083)**  
Processes payments, generates invoices, and manages billing-related operations.
- Payment processing
- Invoice generation
- Dynamic cost calculation
- Payment method management
- Membership-based pricing

![Microservice Architecture](https://github.com/user-attachments/assets/4e1ae943-fa27-4958-a0f4-c4f83f17c3c2)

## Technical Stack

### Backend
- **Language**: Go
- **Database**: PostgreSQL (Supabase)
- **Authentication**: JWT
- **API**: RESTful endpoints

### Frontend
- **Markup Languages**: HTML, CSS, JavaScript
- **Styling**: Tailwind CSS
- **State Management**: Local Storage

## Getting Started

### Dependencies Overview
Ensure these are installed!

- go get -u github.com/golang-jwt/jwt@v3.2.2  **(JWT authentication)**
- go get -u github.com/gorilla/handlers@v1.5.2 **(HTTP middleware)**
- go get -u github.com/gorilla/mux@v1.8.1 **(HTTP router)**
- go get -u github.com/lib/pq@v1.10.9 **(PostgreSQL driver)**
- go get -u golang.org/x/crypto@v0.30.0 **(Cryptographic functions)**

### How to run 
You might need 3 terminals open for this!
1. User Services
```bash
cd services\user-service
go run main.go
```
2. Vehicle Services
```bash
cd services\vehicle-service
go run main.go
```
3. Billing Services
```bash
cd services\billing-service
go run main.go
```

## API Documentation

### User Service Endpoints
```
POST /users/register - Register new user
POST /users/login - User login
PUT /users/{id}/profile - Update user profile
```

### Vehicle Service Endpoints
```
GET /api/vehicles/available - Get available vehicles
POST /api/bookings - Create booking
PUT /api/bookings/{id} - Update booking
DELETE /api/bookings/{id} - Cancel booking
GET /api/bookings/my - Get user bookings
```

### Billing Service Endpoints
```
POST /api/billing/calculate - Calculate rental cost
POST /api/billing/invoices - Create invoice
GET /api/billing/users/{id}/invoices - Get user invoices
POST /api/billing/payment-methods - Add payment method
POST /api/billing/invoices/{id}/pay - Process payment
```

## Database Schema

The system uses a shared PostgreSQL database with separate tables for each service domain:

- Users Table
- Vehicles Table
- Bookings Table
- Invoices Table
- Membership Tier Table
## Security Implementation

### Authentication
- Secure password hashing using bcrypt
- JWT-based authentication with expiration
- Token validation middleware
- Secure HTTP headers

### Data Protection
- Prepared SQL statements
- Input validation
- CORS protection
- Request timeouts

## Scaling Considerations

### Database Connections
- Connection pooling (25 connections per service)
- Connection lifetime management
- Efficient connection handling

### Service Independence
- Independent scaling capability
- Separate frontend assets
- Isolated deployment options

## Monitoring and Maintenance

### Logging
- Structured logging in each service
- Error tracking and monitoring
- Request/response logging

### Health Checks
- Database connection monitoring
- Service status endpoints
- Error rate tracking

## Future Enhancements

1. Service Discovery
   - Implementation of service registry
   - Load balancing capabilities
   - Health check integration

2. Message Queue Integration
   - Asynchronous operations
   - Event-driven architecture
   - Enhanced failure handling

3. Caching Layer
   - Redis integration
   - Cache invalidation strategies
   - Performance optimization

4. Payment Method
   - Credit Card payment
   - Apple Pay payment
   - Other alternatives

## Contact

Your Name - [Casey Tan](https://github.com/MetroBeann)  
Project Link: [https://github.com/MetroBeann/CNAD-CarSharingGO1.git](https://github.com/MetroBeann/CNAD-CarSharingGO1.git)
