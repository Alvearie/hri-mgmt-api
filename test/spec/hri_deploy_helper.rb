# (C) Copyright IBM Corp. 2021
#
# SPDX-License-Identifier: Apache-2.0

class HRIDeployHelper

  def deploy_hri(exe_path, config_path, log_path, log_prefix = '', override_params = nil)
    Open3.popen3("#{exe_path} -config-path=#{config_path} #{override_params} -kafka-properties=security.protocol:sasl_ssl,sasl.mechanism:PLAIN,sasl.username:token,sasl.password:#{ENV['KAFKA_PASSWORD']},ssl.endpoint.identification.algorithm:https 2> #{log_path}/#{log_prefix}error.txt > #{log_path}/#{log_prefix}output.txt &")
    sleep 1
    @error_log = File.read(File.join(log_path, "#{log_prefix}error.txt"))
    @output_log = File.read(File.join(log_path, "#{log_prefix}output.txt"))
    unless @error_log.empty? && !@output_log.include?('"level":"FATAL"')
      raise "A fatal error was encountered when deploying the hri-mgmt-api.
      OUTPUT LOG: #{@output_log}
      ERROR LOG: #{@error_log}"
    end
  end

end