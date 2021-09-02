# (C) Copyright IBM Corp. 2021
#
# SPDX-License-Identifier: Apache-2.0

class HRIDeployHelper

  def deploy_hri(exe_path, config_path, log_path, override_params = nil)
    Open3.popen3("#{exe_path} -config-path=#{config_path} #{override_params} 2> #{log_path}/error.txt > #{log_path}/output.txt &")
    sleep 1
    @error_log = File.read(File.join(File.dirname(__FILE__), 'error.txt'))
    @output_log = File.read(File.join(File.dirname(__FILE__), 'output.txt'))
    unless @error_log.empty? && !@output_log.include?('"level":"FATAL"')
      raise "A fatal error was encountered when deploying the hri-mgmt-api: #{@output_log}"
    end
  end

end