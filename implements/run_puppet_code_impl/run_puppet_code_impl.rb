#! /opt/puppetlabs/puppet/bin/ruby
# frozen_string_literal: true
require 'fileutils'
require 'json'
require 'puppet'
require 'puppet_pal'
require 'puppet/configurer'
require 'securerandom'
require 'tempfile'

# The module dir is expected to exist and already contain modules,
# so no logic below should edit this or use another path
MODULE_DIR = File.join(ENV['HOME'], '.regulator', 'content', 'puppet', 'modules')

def build_program(code)
  ast = Puppet::Pops::Serialization::FromDataConverter.convert(code)
  # Node definitions must be at the top level of the apply block.
  # That means the apply body either a) consists of just a
  # NodeDefinition, b) consists of a BlockExpression which may
  # contain NodeDefinitions, or c) doesn't contain NodeDefinitions.
  # See https://github.com/puppetlabs/bolt/pull/1512 for more details
  definitions = if ast.is_a?(Puppet::Pops::Model::BlockExpression)
                  ast.statements.select { |st| st.is_a?(Puppet::Pops::Model::NodeDefinition) }
                elsif ast.is_a?(Puppet::Pops::Model::NodeDefinition)
                  [ast]
                else
                  []
                end
  # During ordinary compilation, definitions are stored on the parser at
  # parse time and then added to the Program node at the root of the AST
  # before evaluation. Because the AST for an apply block has already been
  # parsed and is not a complete tree with a Program at the root level, we
  # need to rediscover the definitions and construct our own Program object.
  # https://github.com/puppetlabs/bolt/commit/3a7597dda25cdb25854c7d08d37c5c58ab6a016b
  Puppet::Pops::Model::Factory.PROGRAM(ast, definitions, ast.locator).model
end

def compile_code(code, modulepath)
  ast = Puppet::Pops::Parser::EvaluatingParser.new.parse_string(code, nil)
  Puppet::Pal.in_tmp_environment('regulator_catalog', modulepath: [modulepath]) do |pal|
    # This compiler has been configured with a node containing
    # the requested environment, facts, and variables, and is used
    # to compile a catalog in that context from the supplied AST.
    pal.with_catalog_compiler do | compiler |
      Puppet[:strict] = :warning
      Puppet[:strict_variables] = false

      ast = build_program(ast)
      compiler.evaluate(ast)
      compiler.evaluate_ast_node
      compiler.compile_additions
      compiler.catalog_data_hash
    end
  end
end

def setup(noop)
  # Create temporary directories for all core Puppet settings so we don't clobber
  # existing state or read from puppet.conf. Also create a temporary modulepath.
  # Additionally include rundir, which gets its own initialization.
  puppet_root = Dir.mktmpdir
  cli = (Puppet::Settings::REQUIRED_APP_SETTINGS + [:rundir]).flat_map do |setting|
    ["--#{setting}", File.join(puppet_root, setting.to_s.chomp('dir'))]
  end
  cli << '--modulepath' << MODULE_DIR
  Puppet.initialize_settings(cli)

  # Avoid extraneous output
  Puppet[:report] = false

  # Make sure to apply the catalog
  case noop.strip
  when "run"
    Puppet[:noop] = false
  when "observe"
    Puppet[:noop] = true
  else
    $stderr.puts "ERROR: first argument must match 'run' or 'observe', given #{noop.strip}. Cannot continue"
    $stdout.puts "failures"
    exit 1
  end

  Puppet[:default_file_terminus] = :file_server

  # This happens implicitly when running the Configurer, but we make it explicit here. It creates the
  # directories we configured earlier.
  Puppet.settings.use(:main)
  puppet_root
end

def format_report_to_result(raw_report)
  report = raw_report.to_data_hash
  $stderr.puts JSON.pretty_generate(report)

  resource_results = report["metrics"]["resources"]["values"]
  active_results = [
    "restarted",
    "changed",
    "out_of_sync",
    "scheduled",
    "corrective_change",
  ]

  failed_results = [
    "failed",
    "failed_to_restart",
  ]

  result = "conformed"
  resource_results.each do |resource_result|
    if resource_result[2] != 0 && active_results.include?(resource_result[0])
      # Set the result to "changes" but do not return that result yet, as
      # there may be failures that would override the "changes" result
      result = "changes"
    end
    if resource_result[2] != 0 && failed_results.include?(resource_result[0])
      # Failures are the
      result = "failures"
      $stdout.puts result
      return
    end
  end
  $stdout.puts result
end

begin
  puppet_code = ARGV[1]
  puppet_root = setup(ARGV[0])

  env = Puppet.lookup(:environments).get('production')
  # Needed to ensure features are loaded
  env.each_plugin_directory do |dir|
    $LOAD_PATH << dir unless $LOAD_PATH.include?(dir)
  end

  # Ensure custom facts are available for provider suitability tests
  facts = Puppet::Node::Facts.indirection.find(SecureRandom.uuid, environment: env)

  report = if Puppet::Util::Package.versioncmp(Puppet.version, '5.0.0') > 0
             Puppet::Transaction::Report.new
           else
             Puppet::Transaction::Report.new('apply')
           end

  overrides = { current_environment: env,
                loaders: Puppet::Pops::Loaders.new(env) }

  Puppet.override(overrides) do
    raw_compiled_catalog = compile_code(puppet_code, MODULE_DIR)
    catalog = Puppet::Resource::Catalog.from_data_hash(raw_compiled_catalog)
    catalog.environment = env.name.to_s
    catalog.environment_instance = env
    if defined?(Puppet::Pops::Evaluator::DeferredResolver)
      # Only available in Puppet 6
      Puppet::Pops::Evaluator::DeferredResolver.resolve_and_replace(facts, catalog)
    end
    catalog = catalog.to_ral

    configurer = Puppet::Configurer.new
    configurer.run(catalog: catalog, report: report, pluginsync: false)
  end
  format_report_to_result(report)
ensure
  begin
    FileUtils.remove_dir(puppet_root)
  rescue Errno::ENOTEMPTY => e
    $stderr.puts("Could not cleanup temporary directory: #{e}")
  end
end

exit 0
