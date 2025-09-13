// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;

namespace Ycs.Contracts
{
    public class YDocOptions
    {
        private static Predicate<IItem> DefaultPredicate = (item) => true;

        public bool Gc { get; set; } = true;
        public Predicate<IItem> GcFilter { get; set; } = DefaultPredicate;
        public string Guid { get; set; } = System.Guid.NewGuid().ToString("D");
        public IDictionary<string, string> Meta { get; set; } = null;
        public bool AutoLoad { get; set; } = false;

        public YDocOptions Clone()
        {
            return new YDocOptions
            {
                Gc = Gc,
                GcFilter = GcFilter,
                Guid = Guid,
                Meta = Meta == null ? null : new Dictionary<string, string>(Meta),
                AutoLoad = AutoLoad
            };
        }

        public void Write(IUpdateEncoder encoder, int offset)
        {
            var dict = new Dictionary<string, object>();
            dict["gc"] = Gc;
            dict["guid"] = Guid;
            dict["autoLoad"] = AutoLoad;

            if (Meta != null)
            {
                dict["meta"] = Meta;
            }

            encoder.WriteAny(dict);
        }

        public static YDocOptions Read(IUpdateDecoder decoder)
        {
            var dict = (IDictionary<string, object>)decoder.ReadAny();

            var result = new YDocOptions();
            result.Gc = dict.ContainsKey("gc") ? (bool)dict["gc"] : true;
            result.Guid = dict.ContainsKey("guid") ? dict["guid"].ToString() : System.Guid.NewGuid().ToString("D");
            result.Meta = dict.ContainsKey("meta") ? dict["meta"] as Dictionary<string, string> : null;
            result.AutoLoad = dict.ContainsKey("autoLoad") ? (bool)dict["autoLoad"] : false;

            return result;
        }
    }
}
