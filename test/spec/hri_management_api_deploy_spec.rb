require_relative '../env'

describe 'HRI Management API Deploy' do

  before(:all) do
    @request_helper = HRITestHelpers::RequestHelper.new
    @hri_deploy_helper = HRIDeployHelper.new
    @hri_base_url = ENV['HRI_URL']
    @exe_path = File.absolute_path(File.join(File.dirname(__FILE__), "../../src/hri"))
    @log_path = File.absolute_path(File.join(File.dirname(__FILE__), "/"))
    @config_path = File.absolute_path(File.join(File.dirname(__FILE__), "test_config"))
  end

  after(:each) do
    processes = `lsof -iTCP:1323 -sTCP:LISTEN`
    unless processes == ''
      process_id = processes.split("\n").select { |s| s.start_with?('hri') }[0].split(' ')[1].to_i
      `kill #{process_id}` unless process_id.nil?
    end
  end

  after(:all) do
    File.delete("#{@log_path}/output.txt") if File.exists?("#{@log_path}/output.txt")
    File.delete("#{@log_path}/error.txt") if File.exists?("#{@log_path}/error.txt")
  end

  context 'GET /healthcheck' do

    it 'Success Without TLS' do
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml -tls-enabled=false", @log_path)
      response = @request_helper.rest_get("#{@hri_base_url.gsub('https', 'http')}/healthcheck", {})
      raise "Health check failed: #{response.body}" unless response.code == 200
    end

    it 'Success With TLS' do
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
      response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
      raise "Health check failed: #{response.body}" unless response.code == 200
    end

    it 'Invalid Kafka Password' do
      begin
        temp = ENV['KAFKA_PASSWORD']
        ENV['KAFKA_PASSWORD'] = 'INVALID'
        @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
        response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
        expect(response.code).to eq 503
        response_body = JSON.parse(response.body)
        expect(response_body['errorEventId']).to_not be_nil
        expect(response_body['errorDescription']).to eql 'HRI Service Temporarily Unavailable | error Detail: error getting Kafka topics: Local: Broker transport failure'
      ensure
        ENV['KAFKA_PASSWORD'] = temp
      end
    end

    it 'Invalid Kafka Brokers' do
      begin
        temp = ENV['KAFKA_BROKERS']
        ENV['KAFKA_BROKERS'] = 'INVALID:9093'
        @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
        response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
        expect(response.code).to eq 503
        response_body = JSON.parse(response.body)
        expect(response_body['errorEventId']).to_not be_nil
        expect(response_body['errorDescription']).to eql 'HRI Service Temporarily Unavailable | error Detail: error getting Kafka topics: Local: Broker transport failure'
      ensure
        ENV['KAFKA_BROKERS'] = temp
      end
    end

    it 'Invalid Elastic URL' do
      begin
        temp = ENV['ELASTIC_URL']
        ENV['ELASTIC_URL'] = 'INVALID'
        @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
        response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
        expect(response.code).to eq 503
        response_body = JSON.parse(response.body)
        expect(response_body['errorEventId']).to_not be_nil
        expect(response_body['errorDescription']).to eql "Could not perform elasticsearch health check: [500] elasticsearch client error: unsupported protocol scheme \"\""
      ensure
        ENV['ELASTIC_URL'] = temp
      end
    end

    it 'Invalid Elastic Username' do
      begin
        temp = ENV['ELASTIC_USERNAME']
        ENV['ELASTIC_USERNAME'] = 'INVALID'
        @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
        response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
        expect(response.code).to eq 503
        response_body = JSON.parse(response.body)
        expect(response_body['errorEventId']).to_not be_nil
        expect(response_body['errorDescription']).to eql 'Could not perform elasticsearch health check: [500] unexpected Elasticsearch 401 error'
      ensure
        ENV['ELASTIC_USERNAME'] = temp
      end
    end

    it 'Invalid Elastic Password' do
      begin
        temp = ENV['ELASTIC_PASSWORD']
        ENV['ELASTIC_PASSWORD'] = 'INVALID'
        @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
        response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
        expect(response.code).to eq 503
        response_body = JSON.parse(response.body)
        expect(response_body['errorEventId']).to_not be_nil
        expect(response_body['errorDescription']).to eql 'Could not perform elasticsearch health check: [500] unexpected Elasticsearch 401 error'
      ensure
        ENV['ELASTIC_PASSWORD'] = temp
      end
    end

    it 'Invalid Elastic Cert' do
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/invalid_cert_config.yml", @log_path)
      response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
      expect(response.code).to eq 503
      response_body = JSON.parse(response.body)
      expect(response_body['errorEventId']).to_not be_nil
      expect(response_body['errorDescription']).to eql 'Could not perform elasticsearch health check: [500] elasticsearch client error: x509: certificate signed by unknown authority'
    end

  end

  context 'GET /alive' do

    it 'Success' do
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
      response = @request_helper.rest_get("#{@hri_base_url.gsub('/hri', '')}/alive", {})
      raise "Alive check failed: #{response.body}" unless response.code == 200
      expect(response.body).to eql 'yes'
    end

  end

end