# Copyright (c) 2009-2013 VMware, Inc.
# Copyright (c) 2013 Nippon Telegraph and Telephone Corporation

module Bosh::Agent
  class Infrastructure::Cloudstack::Registry
    class << self

      attr_accessor :user_data

      HTTP_API_TIMEOUT     = 300
      HTTP_CONNECT_TIMEOUT = 30
      META_DATA_URI = "http://" +
        %x[grep dhcp-server-identifier /var/lib/dhclient/* /var/lib/dhcp3/* /var/lib/dhcp/* 2>/dev/null | tail -1 | awk '{print $NF}' | tr -d '\;'].strip +
        "/latest"
      USER_DATA_FILE = File.join(File::SEPARATOR, "var", BOSH_APP_USER, "bosh", "user_data.json")

      ##
      # Returns the logger.
      #
      # @return [Logger] Bosh Agent logger
      def logger
        Bosh::Agent::Config.logger
      end

      ##
      # Gets the OpenSSH public key. First we try to get it from the CloudStack meta data endpoint, if we fail,
      # then we fallback to the injected user data file.
      #
      # @return [String] OpenSSH key
      def get_openssh_key
        get_uri(META_DATA_URI + "/public-keys")
      rescue LoadSettingsError => e
        logger.info("Failed to get OpenSSH public key from CloudStack meta data endpoint: #{e.message}")
        user_data = parse_user_data(get_user_data_from_file)
        unless user_data.has_key?("openssh") && user_data["openssh"].has_key?("public_key")
          raise LoadSettingsError, "Cannot get OpenSSH public key from injected user data file: #{user_data.inspect}"
        end
        user_data["openssh"]["public_key"]
      end

      ##
      # Gets the settings for this agent from the Bosh registry.
      #
      # @return [Hash] Agent Settings
      def get_settings
        @registry_endpoint ||= get_registry_endpoint
        url = "#{@registry_endpoint}/instances/#{get_server_name}/settings"
        raw_response = get_uri(url)

        registry_data = Yajl::Parser.parse(raw_response)
        unless registry_data.is_a?(Hash) && registry_data.has_key?("settings")
          raise LoadSettingsError, "Invalid response received from Bosh registry, " \
                                   "got #{registry_data.class}: #{registry_data}"
        end

        settings = Yajl::Parser.parse(registry_data["settings"])
        unless settings.is_a?(Hash)
          raise(LoadSettingsError, "Invalid settings received from Bosh registry, " \
                                   "got #{settings.class}: #{settings}")
        end

        settings
      rescue Yajl::ParseError => e
        raise LoadSettingsError, "Cannot parse settings from Bosh registry, got #{raw_response} - #{e.message}"
      end

      ##
      # Gets the server name from CloudStack user data.
      #
      # @return [String] CloudStack server name
      def get_server_name
        user_data = get_user_data
        unless user_data.has_key?("server") && user_data["server"].has_key?("name")
          raise LoadSettingsError, "Cannot get CloudStack server name from user data #{user_data.inspect}"
        end
        user_data["server"]["name"]
      end

      ##
      # Gets the Bosh registry endpoint from CloudStack user data.
      #
      # @return [String] Bosh registry endpoint
      def get_registry_endpoint
        user_data = get_user_data
        unless user_data.has_key?("registry") && user_data["registry"].has_key?("endpoint")
          raise LoadSettingsError, "Cannot get Bosh registry endpoint from user data #{user_data.inspect}"
        end
        lookup_registry_endpoint(user_data)
      end

      ##
      # If the Bosh registry endpoint is specified with a Bosh DNS name, i.e. 0.registry.default.cloudstack.bosh,
      # then the agent needs to lookup the name and insert the IP address, as the agent doesn't update
      # resolv.conf until after the bootstrap is run.
      #
      # @param [Hash] user_data CloudStack user data (generated by the CPI)
      # @return [String] Bosh registry endpoint
      def lookup_registry_endpoint(user_data)
        registry_endpoint = user_data["registry"]["endpoint"]

        # If user data doesn't contain dns info, there is noting we can do, so just return the endpoint
        return registry_endpoint if user_data["dns"].nil? || user_data["dns"]["nameserver"].nil?

        # If the endpoint is an IP address, just return the endpoint
        registry_hostname = extract_registry_hostname(registry_endpoint)
        return registry_endpoint unless (IPAddr.new(registry_hostname) rescue(nil)).nil?

        nameservers = user_data["dns"]["nameserver"]
        registry_ip = lookup_registry_ip_address(registry_hostname, nameservers)
        inject_registry_ip_address(registry_ip, registry_endpoint)
      rescue Resolv::ResolvError => e
        raise LoadSettingsError, "Cannot lookup #{registry_hostname} using #{nameservers.join(", ")}: #{e.inspect}"
      end

      ##
      # Extracts the hostname from the Bosh registry endpoint.
      #
      # @param [String] endpoint Bosh registry endpoint
      # @return [String] Bosh registry hostname
      def extract_registry_hostname(endpoint)
        match = endpoint.match(%r{https*://([^:]+):})
        unless match && match.size == 2
          raise LoadSettingsError, "Cannot extract Bosh registry hostname from #{endpoint}"
        end
        match[1]
      end

      ##
      # Lookups for the Bosh registry IP address.
      #
      # @param [String] hostname Bosh registry hostname
      # @param [Array] nameservers Array containing nameserver address
      # @return [Resolv::IPv4] Bosh registry IP address
      def lookup_registry_ip_address(hostname, nameservers)
        resolver = Resolv::DNS.new(:nameserver => nameservers)
        resolver.getaddress(hostname)
      end

      ##
      # Injects an IP address in the Bosh registry endpoint.
      #
      # @param [Resolv::IPv4] ip Bosh registry IP address
      # @param [String] endpoint Bosh registry endpoint
      # @return [String] Bosh registry endpoint
      def inject_registry_ip_address(ip, endpoint)
        endpoint.sub(%r{//[^:]+:}, "//#{ip}:")
      end

      ##
      # Gets the CloudStack user data. First we try to get it from the CloudStack user data endpoint, if we fail,
      # then we fallback to the injected user data file.
      #
      # @return [Hash] CloudStack user data
      def get_user_data
        return @user_data if @user_data
        begin
          raw_user_data = get_uri(META_DATA_URI + "/user-data")
        rescue LoadSettingsError => e
          logger.info("Failed to get user data from CloudStack user data endpoint: #{e.message}")
          raw_user_data = get_user_data_from_file
        end

        logger.info("CloudStack user data: #{raw_user_data.inspect}")
        @user_data = parse_user_data(raw_user_data)
      end

      ##
      # Gets the CloudStack user data from the injected user data file.
      #
      # @return [String] CloudStack user data
      def get_user_data_from_file
        File.read(USER_DATA_FILE)
      rescue SystemCallError => e
        raise LoadSettingsError, "Failed to get user data from CloudStack injected user data file: #{e.message}"
      end

      ##
      # Parses the CloudStack user data.
      #
      # @param [String] raw_user_data Raw CloudStack user data
      # @return [Hash] CloudStack user data
      def parse_user_data(raw_user_data)
        begin
          user_data = Yajl::Parser.parse(raw_user_data)
        rescue Yajl::ParseError => e
          raise LoadSettingsError, "Cannot parse user data #{raw_user_data.inspect}: #{e.message}"
        end

        unless user_data.is_a?(Hash)
          raise LoadSettingsError, "Invalid user data format, Hash expected, got #{user_data.class}: #{user_data}"
        end

        user_data
      end

      ##
      # Sends GET request to an specified URI.
      #
      # @param [String] uri URI to request
      # @return [String] Response body
      def get_uri(uri)
        client = HTTPClient.new
        client.send_timeout = HTTP_API_TIMEOUT
        client.receive_timeout = HTTP_API_TIMEOUT
        client.connect_timeout = HTTP_CONNECT_TIMEOUT

        headers = {"Accept" => "application/json"}
        response = client.get(uri, {}, headers)
        unless response.status == 200
          raise LoadSettingsError, "Endpoint #{uri} returned HTTP #{response.status}"
        end

        response.body
      rescue HTTPClient::TimeoutError
        raise LoadSettingsError, "Timed out reading endpoint #{uri}"
      rescue HTTPClient::BadResponseError => e
        raise LoadSettingsError, "Received bad HTTP response from endpoint #{uri}: #{e.inspect}"
      rescue URI::Error, SocketError, Errno::ECONNREFUSED, SystemCallError => e
        raise LoadSettingsError, "Error requesting endpoint #{uri}: #{e.inspect}"
      end
    end
  end
end
