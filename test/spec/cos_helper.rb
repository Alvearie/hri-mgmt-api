# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class COSHelper

  def initialize
    @helper = Helper.new
    @cos_url = ENV['COS_URL']
    @iam_token = IAMHelper.new.get_access_token
  end

  def get_object_data(bucket_name, object_path)
    #Get COS object, or use local data if object does not exist
    response = @helper.rest_get("#{@cos_url}/#{bucket_name}/#{object_path}", {'Authorization' => "Bearer #{@iam_token}"})
    response.code == 200 ? response.body : File.read(File.join(File.dirname(__FILE__), 'fhir_practitioners.txt'))
  end

end