frontend_origin = ENV.fetch('FRONTEND_ORIGIN', 'http://localhost:5173')

Rails.application.config.middleware.insert_before 0, Rack::Cors do
  allow do
    origins frontend_origin
    resource '/api/*',
      headers: :any,
      methods: [:get, :post, :put, :patch, :delete, :options, :head],
      credentials: true
  end
end 