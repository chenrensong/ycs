// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using Ycs.Contracts;

namespace Ycs.Core
{
    /// <summary>
    /// Global accessor for ContentFactory to avoid circular dependencies
    /// This class is initialized by YcsBootstrap
    /// </summary>
    internal static class ContentFactoryAccessor
    {
        private static IContentFactory _factory;

        /// <summary>
        /// Sets the ContentFactory instance (called by YcsBootstrap)
        /// </summary>
        internal static void SetFactory(IContentFactory factory)
        {
            _factory = factory;
        }

        /// <summary>
        /// Gets the ContentFactory instance
        /// </summary>
        internal static IContentFactory Factory
        {
            get
            {
                return _factory ?? throw new System.InvalidOperationException("ContentFactory not initialized. Call YcsBootstrap.Initialize() first.");
            }
        }
    }
}