require "kemal"
require "json"
require "http/client"

# Register with consul
service = Hash{
  "ID" => "projects1",
  "Name" => "projects",
  "Address" => "127.0.0.1",
  "Port" => 3000
  }.to_json
response = HTTP::Client.put("127.0.0.1:8500/v1/agent/service/register", headers: HTTP::Headers{"User-agent" => "AwesomeApp"}, body: service) 

if response.status_code != 200
  puts "Couldn't register service"
end

post "/create" do |env|
  #title = env.params.json["title"].as(String)
  #puts title
  puts env.params.json
  env.response.content_type = "application/json"
  {title: "whatever", id: 1}.to_json
end

Kemal.run

