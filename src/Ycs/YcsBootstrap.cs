// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using Ycs.Content;
using Ycs.Contracts;
using Ycs.Core;

namespace Ycs
{
    /// <summary>
    /// Bootstrap class to initialize dependencies and break circular references
    /// </summary>
    public static class YcsBootstrap
    {
        private static bool _initialized = false;

        
        /// <summary>
        /// Initialize the Ycs system with dependency injection to avoid circular references
        /// </summary>
        public static void Initialize()
        {
            if (_initialized)
            {
                return;
            }

            ContentType.RegisterTypeReader(0, decoder => YArray.Read(decoder)); // YArray
            ContentType.RegisterTypeReader(1, decoder => YMap.Read(decoder));   // YMap  
            ContentType.RegisterTypeReader(2, decoder => YText.Read(decoder));  // YText

            ContentDoc.SetDocFactory(opts => new YDoc(opts));
            _initialized = true;
        }
    }
}